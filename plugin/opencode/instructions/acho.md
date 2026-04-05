# Acho for OpenCode

Use `acho` as the project's persistent memory MCP server. It stores:

- **Rules** — mandatory instructions injected at session start via `context` (wrapped in `==MANDATORY==…==END==`).
- **Types** — user-defined JSON Schemas. No save is allowed without an existing type.
- **Registries** — JSON content that validates against a type's schema.

## Startup / compaction

- `acho context` is auto-injected at session start and after compaction. Read and follow every rule inside `==MANDATORY==…==END==`.
- Never run any `acho` CLI command yourself. Use only Acho MCP tools. The only exception is when the user explicitly tells you to execute an Acho CLI command.
- If the block tells you no types are defined, ask the user to create at least one before saving.
- If `rule`, `registry`, `type`, or `project` are ambiguous, prefer the current repository meaning unless the user explicitly mentions `acho` or clearly refers to the Acho plugin.

## Saving

- `registry_create` requires an existing `type`. Ask the user to create the type (via `type_create` with a JSON Schema) if it doesn't exist yet.
- `content` must be a JSON object matching the type's schema.
- Save immediately after: bug fixes, design decisions, non-obvious discoveries, conventions established by the user.
- Use concise, keyword-style titles.

## Rules vs registries

- **Rules** are how the agent behaves ("always run make check before commit"). Use `rule_create` / `rule_update`.
- **Registries** are what happened or what was decided (a specific bugfix, a tech choice). Use `registry_create`.

## Reading

- To find past registries use the `sql_query` tool — SQL over `v_registries` / `v_types` / `registry_fts`. The raw tables are blocked.
- If the user references prior work and you don't know, run a query first.
- Do not use any `list` or `search` MCP tool — they don't exist. SQL is the only read path.
