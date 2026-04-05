#!/bin/bash
# Acho — SessionStart / PostCompact hook for Claude Code
#
# Prepends a ToolSearch instruction that forces Claude Code to load the
# 11 deferred acho MCP tools, then emits the ==MANDATORY== block with the
# current rules.

if [ "$(acho project status 2>/dev/null)" != "enabled" ]; then
  exit 0
fi

cat <<'TOOLS'
CRITICAL FIRST ACTION — Execute this ToolSearch NOW before doing anything else:
select:mcp__plugin_acho_acho__context,mcp__plugin_acho_acho__registry_create,mcp__plugin_acho_acho__registry_update,mcp__plugin_acho_acho__registry_get,mcp__plugin_acho_acho__registry_delete,mcp__plugin_acho_acho__rule_create,mcp__plugin_acho_acho__rule_update,mcp__plugin_acho_acho__rule_delete,mcp__plugin_acho_acho__type_create,mcp__plugin_acho_acho__type_delete,mcp__plugin_acho_acho__sql_query

After loading the tools, read and apply the ==MANDATORY== block below:
Never run any `acho` CLI command yourself. Use only Acho MCP tools, unless the user explicitly tells you to execute an Acho CLI command.
TOOLS

acho context 2>/dev/null

exit 0
