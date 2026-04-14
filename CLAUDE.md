# Acho MCP

Acho is a persistent memory server for coding agents (Claude Code, Cursor, etc.) built as an MCP server in Go.

It stores **rules** (mandatory instructions loaded at the start of every session), **registry types** (user-defined schemas that constrain what can be saved), and **registries** (structured JSON entries that must match their type's schema). All data lives in SQLite with FTS5; the agent reads via SQL through pre-filtered views.

## Mandatory Rules

- Never add migration code to maintain backwards compatibility with old schemas. If the schema changes, the old database is invalid; users can delete and recreate it.
- Always run `make check` after changes. A task is not done until `make check` passes.

## Architecture

### Tables

- `registries(id, type, title, content, content_flat, project, search_hits, get_hits, update_hits, date)`
  - `content` is a JSON object validated against the schema of its `type` (enforced by `CHECK(json_valid(content))` + service-level schema validation)
  - `content_flat` is a space-separated concatenation of the JSON leaf values; FTS5 indexes this, not `content`
- `rules(id, title, text, project, date)` — free-text rules; injected to agents at session start wrapped in `==MANDATORY==...==END==`
- `registry_types(name, schema, project, date)` — `schema` is a JSON Schema (draft 2020-12, validated via `santhosh-tekuri/jsonschema/v5`)

### Project resolution

All three tables use a single `project` column: empty string (`""`) means global (visible to every project); any other value scopes the row to that specific project. No separate scope field.

For types, `name` is a global PK: one name can only be registered once. At save-time the service resolves which type applies to the current project (project match or global).

### Views (per MCP session)

When an MCP session starts with a detected project, `app.Build` creates temp views:
- `v_registries` — rows visible in this project (`project=P OR project=''`), includes `rowid` for FTS joins
- `v_types` — types applicable in this project

The `sql_query` tool exposes these views only. The raw tables are blocked by a regex blocklist.

### MCP tools

| tool | purpose |
|---|---|
| `rule_create`, `rule_update`, `rule_delete` | manage rules |
| `type_create`, `type_delete` | manage registry types (immutable once created; `force=true` cascades registries on delete) |
| `registry_create`, `registry_update`, `registry_get`, `registry_delete` | registry CRUD; create/update validate content against the resolved type's schema |
| `sql_query` | read-only SQL over `v_registries` / `v_types` / `registry_fts`. Blocks raw tables and writes. |

## Structure

- `cmd/acho/` — CLI entrypoint
- `internal/cli/commands/` — CLI commands (`registries list|get|delete`, `rules list|delete`, `types list|delete`, `stats`, `export`, `import`, `mcp`, `agent-setup`, `project`, `config`, plus hidden internal agent-wiring commands)
- `internal/cli/mcp/` — MCP server and tool registry
- `internal/cli/mcp/tools/` — Individual MCP tools
- `internal/cli/config/` — Configuration loading
- `internal/cli/agent/` — Agent setup (Claude Code, OpenCode)
- `internal/cli/term/` — Terminal theming
- `internal/service/` — StoreService, RuleService, RTypeService
- `internal/persistence/` — Shared types, errors, DB connection (`OpenDB`)
- `internal/persistence/store/` — Registry entity, SQLite repo with FTS5
- `internal/persistence/rule/` — Rule entity, SQLite repo
- `internal/persistence/rtype/` — Registry type entity, SQLite repo

## Commands

All commands run inside Docker via the Makefile. Do not use `go` directly.

```
make build      # Build binary to bin/
make test       # Run tests
make fmt        # Format code
make vet        # Static analysis
make check      # All checks: fmt, vet, build, test
make vendor     # Update vendor dependencies
make shell      # Interactive shell in Go container
make install    # Build and install to ~/.local/bin
make clean      # Remove build artifacts
```
