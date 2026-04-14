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
	env.mustMCP(t, "type_create", `{"name":"decision","schema":"{\"type\":\"object\",\"required\":[\"chose\"],\"properties\":{\"chose\":{\"type\":\"string\"}}}","project":"global"}`)

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
	result := env.runMCP("type_create", `{"name":"rule","schema":"{}","project":"global"}`)
	if !mcpContains(result, "already exists") {
		t.Errorf("expected duplicate error, got %q", result)
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
