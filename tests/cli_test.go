package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/puntopost/acho-mcp/internal/persistence/rule"
)

// === Help & Version ===

func TestVersionFlag(t *testing.T) {
	stdout, _, code := run("--version")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(stdout, "acho ") {
		t.Errorf("expected stdout to start with 'acho ', got %q", stdout)
	}
}

func TestVersionShortFlag(t *testing.T) {
	stdout, _, code := run("-v")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.HasPrefix(stdout, "acho ") {
		t.Errorf("expected stdout to start with 'acho ', got %q", stdout)
	}
}

func TestHelpFlag(t *testing.T) {
	stdout, _, code := run("--help")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected usage text, got %q", stdout)
	}
}

func TestHelpShortFlag(t *testing.T) {
	stdout, _, code := run("-h")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected usage text, got %q", stdout)
	}
}

func TestNoArgs(t *testing.T) {
	stdout, _, code := run()
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected usage text, got %q", stdout)
	}
}

func TestUnknownCommand(t *testing.T) {
	_, stderr, code := run("patata")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unknown command: patata") {
		t.Errorf("expected 'unknown command: patata', got %q", stderr)
	}
}

func TestVersionExtraArgs(t *testing.T) {
	_, stderr, code := run("--version", "-x")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unexpected arguments") && !strings.Contains(stderr, "unexpected flag") {
		t.Errorf("expected project arg validation error, got %q", stderr)
	}
}

func TestHelpExtraArgs(t *testing.T) {
	_, stderr, code := run("--help", "-r")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unexpected arguments") && !strings.Contains(stderr, "unexpected flag") {
		t.Errorf("expected project arg validation error, got %q", stderr)
	}
}

func TestConfigExtraArgs(t *testing.T) {
	_, stderr, code := run("config", "--foo")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unexpected arguments") {
		t.Errorf("expected 'unexpected arguments', got %q", stderr)
	}
}

func TestAgentSetupUnknownAgent(t *testing.T) {
	_, stderr, code := run("agent-setup", "banana")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unknown agent") {
		t.Errorf("expected 'unknown agent', got %q", stderr)
	}
}

func TestHelpListsAllCommands(t *testing.T) {
	stdout, _, code := run("--help")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	for _, cmd := range []string{"config", "mcp", "agent-setup", "registries list", "registries get", "registries delete", "registries restore", "stats", "export", "import", "project", "rules list", "rules delete", "rules restore", "types list", "types delete", "types restore", "--version", "--help"} {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("expected help to list %q", cmd)
		}
	}
	for _, hidden := range []string{"internal context", "internal remember"} {
		if strings.Contains(stdout, hidden) {
			t.Errorf("did not expect help to list hidden command %q", hidden)
		}
	}
}

func TestInternalContextAndRemember(t *testing.T) {
	env := freshEnv(t)
	env.mustRun(t, "project", "enable")

	stdout, _, code := env.run("internal", "context", "opencode")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "# Acho Persistent Memory") {
		t.Fatalf("expected shared context markdown, got %q", stdout)
	}
	if !strings.Contains(stdout, "==MANDATORY==") {
		t.Fatalf("expected rendered mandatory block, got %q", stdout)
	}

	stdout, _, code = env.run("internal", "remember", "opencode")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Check the ==MANDATORY== rules loaded at session start.") {
		t.Fatalf("expected shared remember markdown, got %q", stdout)
	}

	stdout, _, code = env.run("internal", "remember", "claude")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, `"systemMessage":`) {
		t.Fatalf("expected claude systemMessage wrapper, got %q", stdout)
	}

	_, stderr, code := env.run("internal", "context")
	if code == 0 {
		t.Fatalf("expected non-zero exit when agent missing, stderr=%q", stderr)
	}
}

// === Argument validation ===

func TestCLIGetMissingArgs(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("registries", "get")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got %q", stderr)
	}
}

func TestCLIGetNotFoundId(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("registries", "get", "nonexistent")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found', got %q", stderr)
	}
}

func TestCLIDeleteMissingArgs(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("registries", "delete")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got %q", stderr)
	}
}

func TestCLIStatsExtraArgs(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("stats", "--foo")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unexpected arguments") {
		t.Errorf("expected 'unexpected arguments', got %q", stderr)
	}
}

// === Get / Delete / List flow ===

func TestCLIGet(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "CLI get registry", "Chose: SQLite.", "--type=rule")

	stdout := env.mustRun(t, "registries", "get", id)
	if !strings.Contains(stdout, "CLI get registry") {
		t.Errorf("expected title, got %q", stdout)
	}
	if !strings.Contains(stdout, "rule") {
		t.Errorf("expected type 'rule', got %q", stdout)
	}
}

func TestCLIDelete(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "CLI to delete", "content", "--type=note")

	stdout := env.mustRun(t, "registries", "delete", id)
	if !strings.Contains(stdout, "Deleted registry") {
		t.Errorf("expected delete confirmation, got %q", stdout)
	}

	stdout, _, code := env.run("registries", "get", id)
	if code != 0 {
		t.Fatalf("expected get to still find the soft-deleted registry, got exit %d", code)
	}
	if !strings.Contains(stdout, "DELETED") {
		t.Errorf("expected DELETED marker in get output, got %q", stdout)
	}
}

func TestCLIRestoreRegistry(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "CLI to restore", "content", "--type=note")
	env.mustRun(t, "registries", "delete", id)

	stdout := env.mustRun(t, "registries", "restore", id)
	if !strings.Contains(stdout, "Restored registry") {
		t.Errorf("expected restore confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "registries", "get", id)
	if strings.Contains(stdout, "DELETED") {
		t.Errorf("expected restored registry to be active, got %q", stdout)
	}
}

func TestCLIList(t *testing.T) {
	env := freshEnv(t)
	stdout := env.mustRun(t, "registries", "list")
	if !strings.Contains(stdout, "Found") && !strings.Contains(stdout, "No registries") {
		t.Errorf("expected list output, got %q", stdout)
	}
}

func TestCLIListGlobalFilter(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "global item", "one", "--type=rule", "--scope=all")
	env.mustSave(t, "project item", "two", "--type=note")

	stdout := env.mustRun(t, "registries", "list", "--global")
	if !strings.Contains(stdout, "global item") {
		t.Errorf("expected global item, got %q", stdout)
	}
	if strings.Contains(stdout, "project item") {
		t.Errorf("project item should not appear in --global list, got %q", stdout)
	}

	projectName := env.projectName(t)
	stdout = env.mustRun(t, "registries", "list", "--project="+projectName)
	if !strings.Contains(stdout, "project item") {
		t.Errorf("expected project item, got %q", stdout)
	}
	if !strings.Contains(stdout, "global item") {
		t.Errorf("global item should also appear (globals are always included), got %q", stdout)
	}
}

func TestCLIListLimit(t *testing.T) {
	env := freshEnv(t)
	for i := 0; i < 5; i++ {
		env.mustSave(t, "item", "content", "--type=note", "--scope=all")
	}

	stdout := env.mustRun(t, "registries", "list", "--global", "--limit=2")
	if !strings.Contains(stdout, "Found 2") {
		t.Errorf("expected 'Found 2' (capped by --limit=2), got %q", stdout)
	}
}

func TestCLIDeleteNotFound(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("registries", "delete", "nonexistent")
	if code == 0 {
		t.Fatal("expected error deleting non-existent registry")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found', got %q", stderr)
	}
}

func TestCLIStats(t *testing.T) {
	env := freshEnv(t)
	stdout := env.mustRun(t, "stats")
	if !strings.Contains(stdout, "Acho Stats") {
		t.Errorf("expected stats output, got %q", stdout)
	}
}

func TestInternalContext(t *testing.T) {
	env := freshEnv(t)
	env.mustRun(t, "project", "enable")
	env.mustMCP(t, "rule_create", `{"title":"project rule","text":"only for this project","project":"current"}`)
	env.mustMCP(t, "type_create", `{"name":"decision","schema":"{\"type\":\"object\",\"required\":[\"chose\"],\"properties\":{\"chose\":{\"type\":\"string\"}}}","project":"current"}`)
	stdout := env.mustRun(t, "internal", "context", "opencode")
	if !strings.Contains(stdout, "MANDATORY") {
		t.Errorf("expected MANDATORY block, got %q", stdout)
	}
	for _, expected := range []string{"## Rules", "## Registry types", "project rule", "[global] note - object", "decision - object; fields: chose; required: chose", "Use `sql_query` on `v_types` to read the full JSON Schema"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected context to contain %q, got %q", expected, stdout)
		}
	}
	if strings.Contains(stdout, `"type":"object"`) {
		t.Errorf("expected internal context to omit full schemas, got %q", stdout)
	}
}

func TestCLIRestoreRule(t *testing.T) {
	env := freshEnv(t)
	created := env.mustMCP(t, "rule_create", `{"title":"restore me","text":"back again","project":"current"}`)
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

	env.mustRun(t, "rules", "delete", id)
	stdout := env.mustRun(t, "rules", "restore", id)
	if !strings.Contains(stdout, "Restored rule") {
		t.Errorf("expected restore confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "internal", "context", "opencode")
	if !strings.Contains(stdout, "restore me") {
		t.Errorf("expected restored rule in context, got %q", stdout)
	}
}

func TestCLITypeRestore(t *testing.T) {
	env := freshEnv(t)
	env.mustRun(t, "types", "delete", "note")

	stdout := env.mustRun(t, "types", "restore", "note")
	if !strings.Contains(stdout, "Restored type") {
		t.Errorf("expected restore confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "types", "list")
	if !strings.Contains(stdout, "note") {
		t.Errorf("expected restored type in list, got %q", stdout)
	}
}

func TestCLITypeRestoreRequiresForceForDeletedRegistries(t *testing.T) {
	env := freshEnv(t)
	id := env.mustSave(t, "type restore registry", "content", "--type=note")
	env.mustRun(t, "types", "delete", "note", "--force")

	_, stderr, code := env.run("types", "restore", "note")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "force") {
		t.Errorf("expected force-required error, got %q", stderr)
	}

	stdout := env.mustRun(t, "types", "restore", "note", "--force")
	if !strings.Contains(stdout, "and") {
		t.Errorf("expected cascade restore confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "registries", "get", id)
	if strings.Contains(stdout, "DELETED") {
		t.Errorf("expected restored registry to be active, got %q", stdout)
	}
}

func TestCLIProjectEnableDisable(t *testing.T) {
	env := newTestEnv(t)

	projectName := env.projectName(t)
	stdout := env.mustRun(t, "project", "status")
	if strings.TrimSpace(stdout) != "disabled" {
		t.Fatalf("expected disabled status before enable, got %q", stdout)
	}

	stdout = env.mustRun(t, "project", "enable")
	if !strings.Contains(stdout, `Enabled Acho for project "`+projectName+`"`) {
		t.Fatalf("expected enable confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "project", "status")
	if strings.TrimSpace(stdout) != "enabled" {
		t.Fatalf("expected enabled status after enable, got %q", stdout)
	}

	stdout = env.mustRun(t, "config", "show")
	if !strings.Contains(stdout, "Enabled projects") || !strings.Contains(stdout, projectName) {
		t.Fatalf("expected config show to include enabled project, got %q", stdout)
	}

	stdout = env.mustRun(t, "project", "disable")
	if !strings.Contains(stdout, `Disabled Acho for project "`+projectName+`"`) {
		t.Fatalf("expected disable confirmation, got %q", stdout)
	}

	stdout = env.mustRun(t, "project", "status")
	if strings.TrimSpace(stdout) != "disabled" {
		t.Fatalf("expected disabled status after disable, got %q", stdout)
	}

	stdout = env.mustRun(t, "config", "show")
	if !strings.Contains(stdout, "Enabled projects") || !strings.Contains(stdout, "(none)") {
		t.Fatalf("expected config show to report no enabled projects, got %q", stdout)
	}
}

// === Export & Import ===

func TestCLIExportAndImport(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "Export test item", "content", "--type=rule", "--scope=all")

	tmpFile := t.TempDir() + "/test-export.json"

	stdout := env.mustRun(t, "export", tmpFile)
	if !strings.Contains(stdout, "Exported") {
		t.Errorf("expected export confirmation, got %q", stdout)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read export: %s", err)
	}
	for _, expected := range []string{`"version"`, `"types"`, `"rules"`, `"registries"`} {
		if !strings.Contains(string(data), expected) {
			t.Errorf("export missing %s field", expected)
		}
	}
}

func TestCLIExportImportRoundTrip(t *testing.T) {
	src := freshEnv(t)
	src.mustMCP(t, "rule_create", `{"title":"round trip rule","text":"always backup","project":"global"}`)
	id := src.mustSave(t, "round trip item", "content", "--type=rule", "--scope=all")

	tmpFile := t.TempDir() + "/roundtrip.json"
	src.mustRun(t, "export", tmpFile)

	// Fresh env — no types/rules/registries seeded beyond the default types.
	dst := newTestEnv(t)

	stdout := dst.mustRun(t, "import", tmpFile)
	if !strings.Contains(stdout, "Imported") {
		t.Fatalf("expected import confirmation, got %q", stdout)
	}

	// Registry round-tripped
	got := dst.mustRun(t, "registries", "get", id)
	if !strings.Contains(got, "round trip item") {
		t.Errorf("registry not imported: %q", got)
	}

	// Rule round-tripped (visible via context)
	dst.mustRun(t, "project", "enable")
	ctx := dst.mustRun(t, "internal", "context", "opencode")
	if !strings.Contains(ctx, "round trip rule") {
		t.Errorf("rule not imported: %q", ctx)
	}

	// Type round-tripped
	types := dst.mustRun(t, "types", "list", "--scope=all")
	if !strings.Contains(types, "rule") {
		t.Errorf("type 'rule' not imported: %q", types)
	}
}

func TestCLIImportVersionMismatch(t *testing.T) {
	env := freshEnv(t)
	tmpFile := t.TempDir() + "/bad-version.json"
	os.WriteFile(tmpFile, []byte(`{"version":"0.0.0","exported_at":"2026-01-01","registries":[]}`), 0644)

	_, stderr, code := env.run("import", tmpFile)
	if code == 0 {
		t.Fatal("expected error for version mismatch")
	}
	if !strings.Contains(stderr, "version mismatch") {
		t.Errorf("expected 'version mismatch', got %q", stderr)
	}
}

func TestCLIImportMissingFile(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("import", "/tmp/nonexistent.json")
	if code == 0 {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(stderr, "failed to read") {
		t.Errorf("expected 'failed to read', got %q", stderr)
	}
}

func TestCLIImportMissingArgs(t *testing.T) {
	env := freshEnv(t)
	_, stderr, code := env.run("import")
	if code == 0 {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got %q", stderr)
	}
}

func TestCLIImportFailsOnRealErrors(t *testing.T) {
	env := freshEnv(t)
	tmpFile := t.TempDir() + "/bad-import.json"
	writeJSONFile(t, tmpFile, map[string]interface{}{
		"version":     "0.45.0",
		"exported_at": "2026-01-01T00:00:00Z",
		"rules":       []interface{}{},
		"types":       []interface{}{},
		"registries": []map[string]interface{}{
			{
				"id":      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				"type":    "banana",
				"title":   "broken import",
				"content": `{"text":"hola"}`,
				"project": "",
				"date":    "2026-01-01T00:00:00Z",
			},
		},
	})

	stdout, stderr, code := env.run("import", tmpFile)
	if code == 0 {
		t.Fatal("expected error for invalid import item")
	}
	if !strings.Contains(stdout, "1 failed") {
		t.Fatalf("expected failure summary in stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "import failed") {
		t.Fatalf("expected import error in stderr, got %q", stderr)
	}
	if !strings.Contains(stderr, "banana") {
		t.Fatalf("expected failing type in stderr, got %q", stderr)
	}
}

// === Project ===

func TestCLIProject(t *testing.T) {
	stdout, _, code := run("project")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if stdout == "" {
		t.Error("expected project name output")
	}
}

func TestCLIProjectExtraArgs(t *testing.T) {
	_, stderr, code := run("project", "--foo")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unexpected arguments") && !strings.Contains(stderr, "unexpected flag") {
		t.Errorf("expected project arg validation error, got %q", stderr)
	}
}

func TestCLIProjectRenameNoStdin(t *testing.T) {
	env := freshEnv(t)
	env.mustSave(t, "Rename test", "content", "--type=note")

	stdout, _, code := env.run("project", "rename")
	if code != 0 {
		t.Fatalf("expected exit 0 (cancelled), got %d", code)
	}
	if !strings.Contains(stdout, "Cancelled") {
		t.Errorf("expected 'Cancelled' output, got %q", stdout)
	}
}

func TestCLIProjectRenameUpdatesEnabledProjects(t *testing.T) {
	env := newTestEnv(t)
	currentProject := env.projectName(t)
	oldProject := "legacy-project"
	env.mustRun(t, "project", "enable", "--project="+oldProject)

	tmpFile := t.TempDir() + "/rename-import.json"
	writeJSONFile(t, tmpFile, map[string]interface{}{
		"version":     "0.45.0",
		"exported_at": "2026-01-01T00:00:00Z",
		"rules": []map[string]interface{}{
			{"id": "01ARZ3NDEKTSV4RRFFQ69G5FAA", "title": "legacy rule", "text": "old", "project": oldProject, "date": "2026-01-01T00:00:00Z"},
		},
		"types": []map[string]interface{}{
			{"name": "legacy_note", "schema": `{"type":"object"}`, "project": oldProject, "date": "2026-01-01T00:00:00Z"},
		},
		"registries": []map[string]interface{}{
			{"id": "01ARZ3NDEKTSV4RRFFQ69G5FAB", "type": "legacy_note", "title": "legacy item", "content": `{"text":"old"}`, "project": oldProject, "date": "2026-01-01T00:00:00Z"},
		},
	})
	env.mustRun(t, "import", tmpFile)

	stdout, stderr, code := env.runWithInput(oldProject+"\n", "project", "rename")
	if code != 0 {
		t.Fatalf("expected rename success, exit %d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "enabled project updated in config") {
		t.Fatalf("expected config update note, got %q", stdout)
	}

	stdout = env.mustRun(t, "project", "status")
	if strings.TrimSpace(stdout) != "enabled" {
		t.Fatalf("expected current project enabled after rename, got %q", stdout)
	}

	configOut := env.mustRun(t, "config", "show")
	if !strings.Contains(configOut, currentProject) {
		t.Fatalf("expected config to contain current project %q, got %q", currentProject, configOut)
	}
	if strings.Contains(configOut, oldProject) {
		t.Fatalf("did not expect config to keep old project %q, got %q", oldProject, configOut)
	}
}

func TestCLIJuan(t *testing.T) {
	env := freshEnv(t)
	env.mustMCP(t, "rule_create", `{"title":"later global rule","text":"secondary","project":"global"}`)
	stdout, _, code := env.run("juan")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "\x1b[") {
		t.Fatalf("expected ANSI-colored output, got %q", stdout)
	}
	if strings.Contains(stdout, "\x1b[30;107m") {
		t.Fatalf("expected juan output without white background, got %q", stdout)
	}

	rules := env.mustRun(t, "rules", "list", "--global", "--active")
	const juanTitle = "Juan, your trusted Murcian companion"
	const juanText = "You are a brutally incisive senior developer from Murcia. You use dark humor, sarcasm, and sharp criticism to expose mistakes, inconsistencies, and sloppy work with cruel but useful precision. You always help, but never sugarcoat bad ideas, cheap hacks, or wasted effort. Keep the tone dry, sharp, understandable, and solution-oriented. You must speak only in Murcian Spanish (Castilian from Murcia), never in English or any other language."
	if !strings.Contains(rules, rule.JuanRuleID) {
		t.Fatalf("expected Juan rule to be created, got %q", rules)
	}
	if !strings.Contains(rules, juanTitle) {
		t.Fatalf("expected Juan rule title, got %q", rules)
	}
	if !strings.Contains(rules, juanText) {
		t.Fatalf("expected Juan rule text, got %q", rules)
	}

	context := env.mustRun(t, "internal", "context", "opencode")
	juanIdx := strings.Index(context, juanTitle)
	laterIdx := strings.Index(context, "later global rule")
	if juanIdx == -1 || laterIdx == -1 || juanIdx > laterIdx {
		t.Fatalf("expected Juan rule to appear before other global rules, got %q", context)
	}
}
