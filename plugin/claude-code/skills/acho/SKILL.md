---
name: acho
description: "ALWAYS ACTIVE — Persistent memory protocol. Load the MANDATORY block via context at session start; manage rules and types before saving anything; query with sql_query."
---

# Acho Persistent Memory

Acho is your persistent memory across sessions and compactions. It stores three kinds of objects, each in its own table:

- **Rules** — free-text mandatory instructions the agent must follow. Served by `context` wrapped in `==MANDATORY==…==END==`.
- **Registry types** — user-defined JSON Schemas that constrain what can be saved. Nothing can be saved until a type exists.
- **Registries** — JSON objects that must validate against the schema of their type.

All three have a `project` field. Empty string (`""`) means global (applies to every project); any other value scopes the row to that specific project. There is no separate `scope` field.

## Tools

### context — MUST be called at the start of every session
| Param | Required | Description |
|-------|----------|-------------|
| `project` | no | Auto-detected if empty |

Returns the rules visible in the current project (globals + project-specific) wrapped in `==MANDATORY==…==END==`. If no types are defined, the block includes a hint telling you to ask the user for types before saving.

Follow every rule. If rules contradict each other, ask the user how to resolve the inconsistency.

### rule_create — create a new rule
| Param | Required | Description |
|-------|----------|-------------|
| `title` | yes | Short title |
| `text` | yes | Rule text (max 1000 chars) |
| `project` | yes | `"current"` (this project) or `"global"` (every project). |

### rule_update — update an existing rule
| Param | Required | Description |
|-------|----------|-------------|
| `id` | yes | Rule ID to update |
| `title` | yes | New title |
| `text` | yes | New text (max 1000 chars) |
| `project` | yes | `"current"` (this project) or `"global"` (every project). |

### rule_delete — delete a rule
| Param | Required | Description |
|-------|----------|-------------|
| `id` | yes | Rule ID |

### type_create — define a registry type
| Param | Required | Description |
|-------|----------|-------------|
| `name` | yes | Slug `^[a-z][a-z_]*$`, globally unique |
| `schema` | yes | JSON Schema (draft 2020-12) that content must match |
| `project` | yes | `"current"` (this project) or `"global"` (every project). |

Types are **immutable**. To change the schema, delete and recreate.

### type_delete — delete a type
| Param | Required | Description |
|-------|----------|-------------|
| `name` | yes | Type name |
| `force` | no | `true` to cascade-delete registries using this type |

### registry_create — create a new registry
| Param | Required | Description |
|-------|----------|-------------|
| `title` | yes | Short searchable keywords |
| `content` | yes | JSON object that validates against the type's schema |
| `type` | yes | Name of an existing type |
| `project` | yes | `"current"` (this project) or `"global"` (every project). |

### registry_update — update an existing registry
| Param | Required | Description |
|-------|----------|-------------|
| `id` | yes | Registry ID |
| `title` / `content` / `type` | no | Only provided fields are changed |

If `content` is updated it is re-validated against the type's schema.

### registry_get — get full content of a registry
| Param | Required | Description |
|-------|----------|-------------|
| `id` | yes | Registry ID |

### registry_delete — delete a registry
| Param | Required | Description |
|-------|----------|-------------|
| `id` | yes | Registry ID |

### sql_query — read-only SQL search / filter / aggregate
| Param | Required | Description |
|-------|----------|-------------|
| `sql` | yes | A single SELECT (or EXPLAIN/WITH…SELECT) |

Exposed objects:
- `v_registries` — pre-filtered for the current project. Includes `rowid` so you can join with FTS. Columns: `rowid, id, type, title, content, content_flat, project, search_hits, get_hits, update_hits, date`. `project=''` means a global entry. `content` is JSON — use `json_extract(content, '$.field')`.
- `v_types` — pre-filtered types. Columns: `name, schema, project, date`.
- `registry_fts` — FTS5 index. Use `MATCH` and join by `rowid` to `v_registries`. `bm25()` and `snippet()` available.
- `sqlite_master`, `pragma_table_info(...)` — for introspection.

Raw tables (`registries`, `rules`, `registry_types`) are blocked. Rules are not queryable via SQL — they come to you via `context`.

Examples:
```sql
-- Search registries matching 'docker', ranked, with snippet
SELECT r.id, r.title, bm25(registry_fts) AS score,
       snippet(registry_fts, 1, '>>>', '<<<', '...', 10) AS excerpt
FROM registry_fts
JOIN v_registries r ON r.rowid = registry_fts.rowid
WHERE registry_fts MATCH 'docker'
ORDER BY score;

-- Bugs whose cause mentions 'concurrency'
SELECT id, title, json_extract(content, '$.fix') AS fix
FROM v_registries
WHERE type = 'bugfix' AND json_extract(content, '$.cause') LIKE '%concurrency%';

-- What types exist
SELECT name, project, schema FROM v_types ORDER BY name;
```

## Flow — first use of Acho on a project

If the user has not yet defined types, `context` will tell you. Ask the user what kinds of things they want to store (bug fixes, decisions, plans, …) and for each, call `type_create` with an appropriate JSON Schema.

Do not invent schemas silently. Ask the user first.

## When to save — proactively

Save IMMEDIATELY after any of these — do NOT wait to be asked:

- User corrects you or states a preference → usually a **rule** (`rule_create`)
- Architecture or technology choice → a **decision** registry
- Bug is found and fixed → a **bugfix** registry
- Non-obvious discovery → a note-like registry
- User asks to save a plan for later → a **plan** registry

**Before saving**, if no matching type exists, ask the user to approve a schema and call `type_create`.

## When NOT to save

- Trivial info useful only right now
- Things the codebase or git history already documents
- Intermediate steps — only final outcomes
- Reasoning chains or conversation logs
- Unverified assumptions

## Writing registries

- **One idea per registry.** Never bundle topics.
- **Title**: searchable keywords, not a sentence.
- **Content**: JSON matching the type's schema. Concise, declarative.

## MANDATORY

- **Always call `context` at the start of every session.**
- **Never use `acho` CLI via Bash.** Use only MCP tools. Exception: the user explicitly asks.
- Follow every rule in the `==MANDATORY==` block.
- If `rule`, `registry`, `type`, or `project` are ambiguous, prefer the current repository meaning unless the user explicitly mentions `acho` or clearly refers to the Acho plugin.
