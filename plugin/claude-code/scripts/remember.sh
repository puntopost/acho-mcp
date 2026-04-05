#!/bin/bash
# Acho — reminder hook
#
# Fires on UserPromptSubmit and SubagentStop. Tells the model to re-check the
# ==MANDATORY== rules it received at session start, because some may require
# action for the current turn.

if [ "$(acho project status 2>/dev/null)" != "enabled" ]; then
  exit 0
fi

OUTPUT=$(jq -n '{"systemMessage": "Check the ==MANDATORY== rules loaded at session start. Some may require you to act before answering this message (e.g., query sql_query, save a registry, apply a convention)."}')

printf '%s\n' "$OUTPUT"
exit 0
