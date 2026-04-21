package tools

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"vitess.io/vitess/go/vt/sqlparser"
)

var _ Tool = (*sqlQuery)(nil)

func init() {
	RegisterTool(&sqlQuery{})
}

type sqlQuery struct{}

func (t *sqlQuery) Instruction() string {
	return "sql_query — run a read-only SQL query against v_registries, v_types and v_rules (project-filtered)"
}

type SQLQueryInput struct {
	SQL string `json:"sql" jsonschema:"A single SQL statement. Must be SELECT (or EXPLAIN / WITH ... SELECT). See description for allowed tables."`
}

type SQLQueryOutput struct {
	Rows []map[string]interface{} `json:"rows"`
	Info string                   `json:"info,omitempty"`
}

const sqlQueryDescription = `Run a read-only SQL query against the Acho database (SQLite).

ALLOWED:
  - v_registries  project-filtered view of registries; includes rowid.
                  columns: rowid, id, type, title, content, content_flat,
                           project, date
                  project='' means a global entry visible to every project.
                  content is a JSON object; query fields with json_extract(content, '$.field')
                  or the -> / ->> operators. date is a Unix timestamp (integer).
  - v_types       project-filtered view of registry types.
                  columns: name, description, schema, project, date
                  project='' means a global type.
  - v_rules       project-filtered view of rules (mandatory instructions).
                  columns: rowid, id, title, text, project, date
                  project='' means a global rule.
  - registry_fts  FTS5 virtual table over registries. Use with MATCH and join
                  by rowid to v_registries (see example). bm25() and snippet()
                  are available.
  - sqlite_master, pragma_table_info(...)   for schema introspection.

FORBIDDEN:
  - References to the raw tables (registries, rules, registry_types).
  - Any write verb (INSERT, UPDATE, DELETE, DROP, CREATE, ALTER, REPLACE,
    ATTACH, DETACH, VACUUM, REINDEX).
  - PRAGMA statements (use the pragma_table_info function instead).

Examples:

  -- List all registries of a given type
  SELECT id, title, date FROM v_registries WHERE type = 'bugfix' ORDER BY date DESC;

  -- Read a JSON field from content
  SELECT id, json_extract(content, '$.cause') AS cause
  FROM v_registries WHERE type = 'bugfix';

  -- Ranked FTS search with snippet, filtered to current project via v_registries
  SELECT r.id, r.title,
         bm25(registry_fts) AS score,
         snippet(registry_fts, 1, '>>>', '<<<', '...', 10) AS excerpt
  FROM registry_fts
  JOIN v_registries r ON r.rowid = registry_fts.rowid
  WHERE registry_fts MATCH 'docker container'
  ORDER BY score;

  -- Discover columns of a view
  SELECT name, type FROM pragma_table_info('v_registries');

  -- List what types are defined and their schemas
  SELECT name, description, project, schema FROM v_types ORDER BY name;

Results are capped at 500 rows.`

const maxRows = 500

var rawSQLTables = map[string]struct{}{
	"registries":     {},
	"rules":          {},
	"registry_types": {},
}

var (
	reFallbackWriteStart = regexp.MustCompile(`(?i)^\s*(insert|update|delete|drop|create|alter|replace|attach|detach|vacuum|reindex|truncate)\b`)
	reFallbackPragmaStmt = regexp.MustCompile(`(?i)^\s*pragma\b`)
	reFallbackRawSource  = regexp.MustCompile(`(?i)\b(from|join)\s+(registries|rules|registry_types)\b`)
	reFallbackExplainQP  = regexp.MustCompile(`(?i)^\s*explain\s+query\s+plan\s+select\b`)
	reFallbackPragmaTF   = regexp.MustCompile(`(?i)^\s*select\b.*\bpragma_table_info\s*\(`)
	reFallbackJSONTable  = regexp.MustCompile(`(?i)^\s*select\b.*\b(json_each|json_tree)\s*\(`)
	reFallbackFTSMatch   = regexp.MustCompile(`(?i)^\s*select\b.*\bmatch\b`)
)

func (t *sqlQuery) Register(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sql_query",
		Description: sqlQueryDescription,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SQLQueryInput) (*mcp.CallToolResult, SQLQueryOutput, error) {
		start := logToolStart("sql_query", "sql_len", len(input.SQL))
		if err := validateSQL(input.SQL); err != nil {
			logToolError("sql_query", start, err)
			return nil, SQLQueryOutput{}, fmt.Errorf("sql_query rejected: %w", err)
		}

		qctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		rows, err := deps.DB.QueryContext(qctx, input.SQL)
		if err != nil {
			logToolError("sql_query", start, err)
			return nil, SQLQueryOutput{}, fmt.Errorf("sql_query failed: %w", err)
		}
		defer rows.Close()

		result, truncated, err := scanRows(rows, maxRows)
		if err != nil {
			logToolError("sql_query", start, err)
			return nil, SQLQueryOutput{}, fmt.Errorf("sql_query scan failed: %w", err)
		}

		logToolSuccess("sql_query", start, "row_count", len(result), "truncated", truncated)
		out := SQLQueryOutput{Rows: result}
		if truncated {
			out.Info = fmt.Sprintf("results truncated at %d rows; add LIMIT or refine the query", maxRows)
		}
		return nil, out, nil
	})
}

func validateSQL(s string) error {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return fmt.Errorf("sql is required")
	}
	err := validateSQLWithParser(trimmed)
	if err == nil {
		return nil
	}

	normalized, normErr := normalizeSQLForValidation(trimmed)
	if normErr != nil {
		return normErr
	}
	if reFallbackWriteStart.MatchString(normalized) {
		return fmt.Errorf("write statements are not allowed")
	}
	if reFallbackPragmaStmt.MatchString(normalized) {
		return fmt.Errorf("PRAGMA statements are not allowed; use pragma_table_info(name) as a table function instead")
	}
	if reFallbackRawSource.MatchString(normalized) {
		return fmt.Errorf("access to raw tables (registries, rules, registry_types) is not allowed; use v_registries / v_types")
	}
	if allowsSQLiteReadFallback(normalized) {
		return nil
	}
	return err
}

func validateSQLWithParser(sql string) error {
	parser := sqlparser.NewTestParser()
	stmt, err := parser.Parse(sql)
	if err != nil {
		return err
	}

	if err := validateSQLStatement(stmt); err != nil {
		return err
	}

	cteNames := collectCTENames(stmt)
	for _, table := range collectReferencedTables(stmt) {
		name := strings.ToLower(table.Name.String())
		if _, ok := cteNames[name]; ok {
			continue
		}
		if _, forbidden := rawSQLTables[name]; forbidden {
			return fmt.Errorf("access to raw tables (registries, rules, registry_types) is not allowed; use v_registries / v_types")
		}
	}

	return nil
}

func validateSQLStatement(stmt sqlparser.Statement) error {
	switch node := stmt.(type) {
	case *sqlparser.Select, *sqlparser.Union:
		return nil
	case *sqlparser.ExplainStmt:
		if node.Statement == nil {
			return fmt.Errorf("EXPLAIN must wrap a statement")
		}
		return validateSQLStatement(node.Statement)
	default:
		return fmt.Errorf("only SELECT, EXPLAIN, or WITH ... SELECT statements are allowed")
	}
}

func collectCTENames(stmt sqlparser.Statement) map[string]struct{} {
	names := make(map[string]struct{})
	_ = sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
		cte, ok := node.(*sqlparser.CommonTableExpr)
		if !ok {
			return true, nil
		}
		names[strings.ToLower(cte.ID.String())] = struct{}{}
		return true, nil
	}, stmt)
	return names
}

func collectReferencedTables(stmt sqlparser.Statement) []sqlparser.TableName {
	seen := make(map[string]struct{})
	out := make([]sqlparser.TableName, 0)
	_ = sqlparser.Walk(func(node sqlparser.SQLNode) (bool, error) {
		table, ok := node.(sqlparser.TableName)
		if !ok {
			return true, nil
		}
		if table.IsEmpty() {
			return true, nil
		}
		name := strings.ToLower(table.Qualifier.String() + "." + table.Name.String())
		if _, exists := seen[name]; exists {
			return true, nil
		}
		seen[name] = struct{}{}
		out = append(out, table)
		return true, nil
	}, stmt)
	return out
}

func allowsSQLiteReadFallback(normalized string) bool {
	return reFallbackExplainQP.MatchString(normalized) ||
		reFallbackPragmaTF.MatchString(normalized) ||
		reFallbackJSONTable.MatchString(normalized) ||
		reFallbackFTSMatch.MatchString(normalized)
}

func normalizeSQLForValidation(s string) (string, error) {
	var out strings.Builder
	state := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		next := byte(0)
		if i+1 < len(s) {
			next = s[i+1]
		}

		switch state {
		case 0:
			switch {
			case ch == '\'' || ch == '"' || ch == '`':
				state = ch
				out.WriteByte(' ')
			case ch == '[':
				state = ']'
				out.WriteByte(' ')
			case ch == '-' && next == '-':
				state = 'n'
				out.WriteByte(' ')
				i++
			case ch == '/' && next == '*':
				state = 'b'
				out.WriteByte(' ')
				i++
			default:
				out.WriteByte(ch)
			}
		case '\'', '"', '`':
			if ch == state {
				if next == state {
					i++
					continue
				}
				state = 0
			}
		case ']':
			if ch == ']' {
				state = 0
			}
		case 'n':
			if ch == '\n' {
				state = 0
				out.WriteByte('\n')
			}
		case 'b':
			if ch == '*' && next == '/' {
				state = 0
				i++
			}
		}
	}
	if state == '\'' || state == '"' || state == '`' || state == ']' || state == 'b' {
		return "", fmt.Errorf("sql has an unterminated string or comment")
	}
	return out.String(), nil
}

func scanRows(rows *sql.Rows, limit int) ([]map[string]interface{}, bool, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}
	out := make([]map[string]interface{}, 0)
	truncated := false
	for rows.Next() {
		if len(out) >= limit {
			truncated = true
			break
		}
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, false, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, c := range cols {
			row[c] = normalizeValue(vals[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return out, truncated, nil
}

func normalizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case []byte:
		return string(x)
	default:
		return x
	}
}
