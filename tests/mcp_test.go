package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mcpSavePayload builds a save payload. `scope` is a legacy test argument;
// scope="all" translates to project="" (global); other scopes are ignored
// because the new design has no scope field.
func mcpSavePayload(title, content, typ, scope string) string {
	wrapped := wrapAsJSON(content)
	args := map[string]interface{}{
		"title":   title,
		"content": wrapped,
		"type":    typ,
		"project": "current",
	}
	if scope == "all" {
		args["project"] = "global"
	}
	b, _ := json.Marshal(args)
	return string(b)
}

// === MCP Save, Search, Get, Update, Delete flow ===

func TestMCPSave(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "registry_create", mcpSavePayload("MCP test rule", "Do: test. Dont: skip.", "rule", "all"))
	if !mcpContains(result, "Created") {
		t.Errorf("expected create confirmation, got %q", result)
	}
}

func TestMCPSaveAndGet(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "MCP get test", "Full content for MCP get", "--type=note")

	result := env.mustMCP(t, "registry_get", fmt.Sprintf(`{"id":"%s"}`, id))
	if !mcpContains(result, "MCP get test") {
		t.Errorf("expected 'MCP get test', got %q", result)
	}
	if !mcpContains(result, "Full content for MCP get") {
		t.Errorf("expected full content, got %q", result)
	}
}

func TestMCPUpdate(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "MCP update test", "Original content", "--type=note")

	result := env.mustMCP(t, "registry_update", fmt.Sprintf(`{"id":"%s","title":"MCP updated title"}`, id))
	if !mcpContains(result, "Updated") {
		t.Errorf("expected update confirmation, got %q", result)
	}

	stdout := env.mustRun(t, "registries", "get", id)
	if !strings.Contains(stdout, "MCP updated title") {
		t.Errorf("expected 'MCP updated title', got %q", stdout)
	}
}

func TestMCPDisabledProjectStartsEmpty(t *testing.T) {
	env := newTestEnv(t)
	result, stderr, code := env.listMCPTools()
	if code != 0 {
		t.Fatalf("expected disabled project to start cleanly, got exit %d, stderr %q", code, stderr)
	}
	if !strings.Contains(result, `"tools":[]`) {
		t.Fatalf("expected empty tools list when disabled, got %q", result)
	}
}

func TestRuleCreateReturnsMandatoryBlock(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "rule_create", `{"title":"fresh rule","text":"obey this now","project":"global"}`)
	if !mcpContains(result, "==MANDATORY==") {
		t.Fatalf("expected mandatory block, got %q", result)
	}
	if !mcpContains(result, "fresh rule") || !mcpContains(result, "obey this now") {
		t.Fatalf("expected created rule contents in response, got %q", result)
	}
	if !mcpContains(result, "==END==") {
		t.Fatalf("expected mandatory end marker, got %q", result)
	}
}

func TestRuleUpdateReturnsReplacementMandatoryBlock(t *testing.T) {
	env := freshEnv(t)
	created := env.mustMCP(t, "rule_create", `{"title":"mutable rule","text":"do the old thing","project":"global"}`)
	idx := strings.Index(created, `"id":"`)
	if idx == -1 {
		t.Fatalf("expected id in create response, got %q", created)
	}
	rest := created[idx+6:]
	end := strings.Index(rest, `"`)
	if end == -1 {
		t.Fatalf("expected closing quote for id, got %q", created)
	}
	id := rest[:end]

	result := env.mustMCP(t, "rule_update", fmt.Sprintf(`{"id":"%s","text":"do the new thing"}`, id))
	if !mcpContains(result, "==MANDATORY==") {
		t.Fatalf("expected mandatory block, got %q", result)
	}
	if !mcpContains(result, "Stop following the previous version of this rule") {
		t.Fatalf("expected replacement instruction, got %q", result)
	}
	if !mcpContains(result, "mutable rule") || !mcpContains(result, "do the new thing") {
		t.Fatalf("expected updated rule contents in response, got %q", result)
	}
	if !mcpContains(result, "==END==") {
		t.Fatalf("expected mandatory end marker, got %q", result)
	}
}

// === MCP Validation ===

func TestMCPSaveUnknownType(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("registry_create", mcpSavePayload("test", "test", "banana", ""))
	if result == "" {
		t.Fatal("no MCP response")
	}
	if !mcpContains(result, "not found") && !mcpContains(result, "banana") {
		t.Errorf("expected type-not-found error, got %q", result)
	}
}

func TestMCPDeleteNotFound(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("registry_delete", `{"id":"nonexistent"}`)
	if result == "" {
		t.Fatal("no MCP response")
	}
	if !mcpContains(result, "not found") {
		t.Errorf("expected 'not found' error, got %q", result)
	}
}

func TestMCPGetNotFound(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("registry_get", `{"id":"nonexistent"}`)
	if result == "" {
		t.Fatal("no MCP response")
	}
	if !mcpContains(result, "not found") {
		t.Errorf("expected 'not found' error, got %q", result)
	}
}

func TestMCPDelete(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "MCP deletable item", "Temporary", "--type=note")

	result := env.mustMCP(t, "registry_delete", fmt.Sprintf(`{"id":"%s"}`, id))
	if !mcpContains(result, "Deleted") {
		t.Errorf("expected delete confirmation, got %q", result)
	}

	stdout, _, code := env.run("registries", "get", id)
	if code != 0 {
		t.Errorf("expected get to still find the soft-deleted registry, got exit %d", code)
	}
	if !strings.Contains(stdout, "DELETED") {
		t.Errorf("expected get output to mark the registry as DELETED, got %q", stdout)
	}
}

// === Types ===

func TestTypeCreateAndUseViaMCP(t *testing.T) {
	env := freshEnv(t)
	env.mustMCP(t, "type_create", `{"name":"decision","description":"Project decision records","schema":"{\"type\":\"object\",\"required\":[\"chose\"],\"properties\":{\"chose\":{\"type\":\"string\"}}}","project":"global"}`)

	// valid content
	env.mustMCP(t, "registry_create", `{"title":"pick db","content":"{\"chose\":\"sqlite\"}","type":"decision","project":"global"}`)

	// invalid content (missing required field) should fail
	result := env.runMCP("registry_create", `{"title":"bad","content":"{}","type":"decision","project":"global"}`)
	if !mcpContains(result, "does not match schema") && !mcpContains(result, "required") {
		t.Errorf("expected schema validation error, got %q", result)
	}
}

func TestTypeCreateDuplicate(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("type_create", `{"name":"rule","description":"duplicate test","schema":"{}","project":"global"}`)
	if !mcpContains(result, "already exists") {
		t.Errorf("expected duplicate error, got %q", result)
	}
}

func TestTypeCreateDescriptionTooLong(t *testing.T) {
	env := freshEnv(t)
	description := strings.Repeat("a", 301)
	result := env.runMCP("type_create", fmt.Sprintf(`{"name":"playbook","description":"%s","schema":"{}","project":"global"}`, description))
	if !mcpContains(result, "description too long") {
		t.Errorf("expected description length error, got %q", result)
	}
}

func TestTypeDeleteWithRegistriesRequiresForce(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "an item", "content", "--type=note")

	result := env.runMCP("type_delete", `{"name":"note"}`)
	if !mcpContains(result, "force") {
		t.Errorf("expected force-required error, got %q", result)
	}

	env.mustMCP(t, "type_delete", `{"name":"note","force":true}`)
}

func TestTypeRenameUpdatesRegistries(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "an item", "content", "--type=note")
	env.mustSave(t, "another", "more content", "--type=note")

	out := env.mustMCP(t, "type_rename", `{"old_name":"note","new_name":"memo"}`)
	if !mcpContains(out, "2 registries updated") {
		t.Errorf("expected 2 registries updated, got %q", out)
	}

	rows := env.mustMCP(t, "sql_query", `{"sql":"SELECT COUNT(*) AS n FROM v_registries WHERE type = 'memo'"}`)
	if !mcpContains(rows, `"n":2`) {
		t.Errorf("expected both registries to have type=memo, got %q", rows)
	}
	typesRows := env.mustMCP(t, "sql_query", `{"sql":"SELECT name FROM v_types WHERE name IN ('note','memo')"}`)
	if mcpContains(typesRows, `"name":"note"`) || !mcpContains(typesRows, `"name":"memo"`) {
		t.Errorf("expected type renamed, got %q", typesRows)
	}
}

func TestTypeRenameCollisionFails(t *testing.T) {
	env := freshEnv(t)
	// `note` and `plan` are both seeded — renaming onto an existing name must fail.
	result := env.runMCP("type_rename", `{"old_name":"note","new_name":"plan"}`)
	if !mcpContains(result, "already exists") {
		t.Errorf("expected collision error, got %q", result)
	}
}

func TestTypeRenameUnknownFails(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("type_rename", `{"old_name":"ghost","new_name":"phantom"}`)
	if !mcpContains(result, "not found") && !mcpContains(result, "ghost") {
		t.Errorf("expected not-found error, got %q", result)
	}
}

func TestTypeRenameInvalidName(t *testing.T) {
	env := freshEnv(t)
	result := env.runMCP("type_rename", `{"old_name":"note","new_name":"Bad-Name"}`)
	if !mcpContains(result, "invalid") {
		t.Errorf("expected invalid name error, got %q", result)
	}
}

// === sql_query ===

func TestSQLQueryOnViewReturnsRows(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "sql item one", "hello world", "--type=note", "--scope=all")
	env.mustSave(t, "sql item two", "another thing", "--type=note", "--scope=all")

	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT title FROM v_registries ORDER BY title"}`)
	if !mcpContains(result, "sql item one") || !mcpContains(result, "sql item two") {
		t.Errorf("expected both items, got %q", result)
	}
}

func TestSQLQueryBlocksRawTables(t *testing.T) {
	env := freshEnv(t)
	for _, q := range []string{
		`{"sql":"SELECT * FROM registries"}`,
		`{"sql":"SELECT * FROM rules"}`,
		`{"sql":"SELECT * FROM registry_types"}`,
	} {
		result := env.runMCP("sql_query", q)
		if !mcpContains(result, "raw tables") {
			t.Errorf("expected raw-tables rejection for %s, got %q", q, result)
		}
	}
}

func TestSQLQueryAllowsAliasNamedLikeRawTable(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT COUNT(*) AS rules FROM v_rules"}`)
	if !mcpContains(result, "rules") {
		t.Fatalf("expected alias named like raw table to be allowed, got %q", result)
	}
}

func TestSQLQueryAllowsRawTableNameInStringLiteral(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT 'registry_types' AS label FROM v_types"}`)
	if !mcpContains(result, "registry_types") {
		t.Fatalf("expected raw table name in string literal to be allowed, got %q", result)
	}
}

func TestSQLQueryAllowsCTEName(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"WITH memo AS (SELECT name FROM v_types) SELECT name FROM memo"}`)
	if !mcpContains(result, "note") {
		t.Fatalf("expected cte query to be allowed, got %q", result)
	}
}

func TestSQLQueryAllowsJSONEach(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT value FROM json_each('[1,2]') ORDER BY value"}`)
	if !mcpContains(result, `"value":1`) || !mcpContains(result, `"value":2`) {
		t.Fatalf("expected json_each result, got %q", result)
	}
}

func TestSQLQueryAllowsExplainQueryPlan(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"EXPLAIN QUERY PLAN SELECT name FROM v_types"}`)
	if !mcpContains(result, "detail") {
		t.Fatalf("expected explain query plan result, got %q", result)
	}
}

func TestSQLQueryBlocksWrites(t *testing.T) {
	env := freshEnv(t)
	for _, q := range []string{
		`{"sql":"DELETE FROM v_registries"}`,
		`{"sql":"UPDATE v_registries SET title='x'"}`,
		`{"sql":"DROP VIEW v_registries"}`,
		`{"sql":"ATTACH DATABASE 'x' AS other"}`,
	} {
		result := env.runMCP("sql_query", q)
		if !mcpContains(result, "not allowed") {
			t.Errorf("expected write rejection for %s, got %q", q, result)
		}
	}
}

func TestSQLQueryFTSWithRowidJoin(t *testing.T) {
	env := freshEnv(t)
	env.mustMCP(t, "registry_create", mcpSavePayload("docker notes", "how to run docker containers", "note", "all"))
	env.mustMCP(t, "registry_create", mcpSavePayload("unrelated", "nothing here", "note", "all"))

	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT r.title FROM registry_fts JOIN v_registries r ON r.rowid = registry_fts.rowid WHERE registry_fts MATCH 'docker'"}`)
	if !mcpContains(result, "docker notes") {
		t.Errorf("expected FTS result, got %q", result)
	}
	if mcpContains(result, "unrelated") {
		t.Errorf("unrelated should not match, got %q", result)
	}
}

func TestSQLQueryAllowsPragmaTableInfoWithRawTableLiteral(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT name FROM pragma_table_info('registry_types') ORDER BY name"}`)
	if !mcpContains(result, "schema") || !mcpContains(result, "project") {
		t.Fatalf("expected pragma_table_info result, got %q", result)
	}
}

func TestSQLQueryAllowsSqliteMasterLiteral(t *testing.T) {
	env := freshEnv(t)
	result := env.mustMCP(t, "sql_query", `{"sql":"SELECT name FROM sqlite_master WHERE name = 'rules'"}`)
	if !mcpContains(result, "rules") {
		t.Fatalf("expected sqlite_master result, got %q", result)
	}
}
