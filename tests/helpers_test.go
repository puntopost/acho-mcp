package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const binary = "../bin/acho"

func binaryPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(binary)
	if err != nil {
		t.Fatalf("failed to resolve binary path: %s", err)
	}
	return path
}

type testEnv struct {
	dir     string
	workdir string
}

// freshEnv creates an isolated ACHO_PATH and seeds a set of permissive types
// (`rule`, `bugfix`, `note`, `plan`, `resource`) so legacy-style tests can
// `save --type=X` without worrying about schemas.
func freshEnv(t *testing.T) *testEnv {
	t.Helper()
	e := newTestEnv(t)
	e.mustRun(t, "project", "enable")
	for _, typ := range []string{"rule", "bugfix", "note", "plan", "resource"} {
		e.seedType(t, typ, typ+" helper type", `{"type":"object"}`)
	}
	return e
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	root := t.TempDir()
	workdir := filepath.Join(root, "test-project")
	if err := os.MkdirAll(workdir, 0755); err != nil {
		t.Fatalf("failed to create workdir: %s", err)
	}
	return &testEnv{dir: root, workdir: workdir}
}

func (e *testEnv) seedType(t *testing.T, name, description, schema string) {
	t.Helper()
	payload, _ := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
		"schema":      schema,
		"project":     "global",
	})
	result := e.runMCP("type_create", string(payload))
	if result == "" {
		t.Fatalf("seedType(%s): no MCP response", name)
	}
	if mcpHasError(result) {
		t.Fatalf("seedType(%s) failed: %s", name, result)
	}
}

func run(args ...string) (string, string, int) {
	path, _ := filepath.Abs(binary)
	cmd := exec.Command(path, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	}

	return stdout.String(), stderr.String(), code
}

func (e *testEnv) run(args ...string) (string, string, int) {
	return e.runWithInput("", args...)
}

func (e *testEnv) runWithInput(input string, args ...string) (string, string, int) {
	path, _ := filepath.Abs(binary)
	cmd := exec.Command(path, args...)
	cmd.Env = append(os.Environ(), "ACHO_PATH="+e.dir)
	cmd.Dir = e.workdir
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	}

	return stdout.String(), stderr.String(), code
}

func (e *testEnv) runMCP(toolName string, argsJSON string) string {
	result, _, _ := e.runMCPDetailed(toolName, argsJSON)
	return result

}

func (e *testEnv) listMCPTools() (string, string, int) {
	input := "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{},\"clientInfo\":{\"name\":\"test\",\"version\":\"1.0\"}}}\n" +
		"{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n" +
		"{\"jsonrpc\":\"2.0\",\"id\":99,\"method\":\"tools/list\"}\n"
	return e.runMCPInput(input)
}

func (e *testEnv) runMCPDetailed(toolName string, argsJSON string) (string, string, int) {
	input := fmt.Sprintf(
		"{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{},\"clientInfo\":{\"name\":\"test\",\"version\":\"1.0\"}}}\n"+
			"{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n"+
			"{\"jsonrpc\":\"2.0\",\"id\":99,\"method\":\"tools/call\",\"params\":{\"name\":\"%s\",\"arguments\":%s}}\n",
		toolName, argsJSON)
	return e.runMCPInput(input)
}

func (e *testEnv) runMCPInput(input string) (string, string, int) {

	path, _ := filepath.Abs(binary)
	cmd := exec.Command(path, "mcp")
	cmd.Env = append(os.Environ(), "ACHO_PATH="+e.dir)
	cmd.Dir = e.workdir

	stdinPipe, _ := cmd.StdinPipe()
	stdoutPipe, _ := cmd.StdoutPipe()
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return "", stderr.String(), 1
	}
	stdinPipe.Write([]byte(input))

	scanner := bufio.NewScanner(stdoutPipe)
	result := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, `"id":99`) {
			result = line
			break
		}
	}

	stdinPipe.Close()
	err = cmd.Wait()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		code = 1
	}

	return result, stderr.String(), code
}

func writeJSONFile(t *testing.T, path string, v interface{}) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write json file: %v", err)
	}
}

func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	return string(b)
}

func (e *testEnv) mustRun(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, code := e.run(args...)
	if code != 0 {
		t.Fatalf("%s failed (exit %d): %s", args[0], code, stderr)
	}
	return stdout
}

// mustSave invokes the MCP `save` tool. Arguments mimic the old CLI save:
// positional title and content, then flags --type=X and --scope=Y (legacy;
// scope=all becomes project=""). Content is wrapped as {"text":"..."} if not
// already valid JSON.
func (e *testEnv) mustSave(t *testing.T, args ...string) string {
	t.Helper()
	if len(args) < 2 {
		t.Fatalf("mustSave needs at least title and content, got %d args", len(args))
	}
	title := args[0]
	content := wrapAsJSON(args[1])
	typ := ""
	scope := ""
	for _, a := range args[2:] {
		switch {
		case strings.HasPrefix(a, "--type="):
			typ = strings.TrimPrefix(a, "--type=")
		case strings.HasPrefix(a, "--scope="):
			scope = strings.TrimPrefix(a, "--scope=")
		}
	}
	if typ == "" {
		typ = "note"
	}
	payload := map[string]interface{}{
		"title":   title,
		"content": content,
		"type":    typ,
		"project": "current",
	}
	if scope == "all" {
		payload["project"] = "global"
	}
	b, _ := json.Marshal(payload)
	result := e.mustMCP(t, "registry_create", string(b))
	// extract id from JSON response
	idx := strings.Index(result, `"id":"`)
	if idx < 0 {
		t.Fatalf("save succeeded but no id in response: %s", result)
	}
	rest := result[idx+6:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		t.Fatalf("malformed id in response: %s", result)
	}
	return rest[:end]
}

func wrapAsJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err == nil {
		return s
	}
	b, _ := json.Marshal(map[string]string{"text": s})
	return string(b)
}

func (e *testEnv) mustMCP(t *testing.T, toolName string, argsJSON string) string {
	t.Helper()
	result := e.runMCP(toolName, argsJSON)
	if result == "" {
		t.Fatalf("no MCP response for %s", toolName)
	}
	if mcpHasError(result) {
		t.Fatalf("%s failed: %s", toolName, result)
	}
	return result
}

func (e *testEnv) projectName(t *testing.T) string {
	t.Helper()
	stdout := e.mustRun(t, "project")
	return strings.TrimSpace(stdout)
}

func extractID(saveOutput string) string {
	parts := strings.Split(saveOutput, "#")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimRight(strings.Fields(parts[1])[0], ":")
}

func mcpContains(result string, text string) bool {
	return strings.Contains(result, text)
}

func mcpHasError(result string) bool {
	return strings.Contains(result, `"isError":true`) || strings.Contains(result, `"error"`)
}
