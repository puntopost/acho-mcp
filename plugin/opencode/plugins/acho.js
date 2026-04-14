async function inject(client, sessionID, text) {
  if (!text || !text.trim()) {
    return
  }

  await client.session.prompt({
    path: { id: sessionID },
    body: {
      noReply: true,
      parts: [{ type: "text", text }],
    },
  })
}

async function achoContext($) {
  try {
    const output = await $`acho internal context opencode`.quiet()
    return output.text().trim()
  } catch {
    return ""
  }
}

async function achoEnabled($) {
  try {
    const output = await $`acho project status`.quiet()
    return output.text().trim() === "enabled"
  } catch {
    return false
  }
}

async function rememberText($) {
  try {
    const output = await $`acho internal remember opencode`.quiet()
    return output.text().trim()
  } catch {
    return ""
  }
}

async function remember(client, $, sessionID) {
  await inject(client, sessionID, await rememberText($))
}

function todoStatus(value) {
  return typeof value === "string" ? value.toLowerCase() : ""
}

function todoCompleted(properties) {
  if (!properties || typeof properties !== "object") {
    return false
  }

  const current =
    properties.todo ?? properties.item ?? properties.current ?? properties.after ?? properties.next
  const previous =
    properties.previous ?? properties.before ?? properties.prior ?? properties.old

  const currentStatus = todoStatus(current?.status)
  const previousStatus = todoStatus(previous?.status)

  return currentStatus === "completed" && previousStatus !== "completed"
}

export const AchoPlugin = async ({ client, $ }) => {
  return {
    event: async ({ event }) => {
      if (!(await achoEnabled($))) {
        return
      }

      if (event.type === "session.created") {
        const context = await achoContext($)
        await inject(client, event.properties.info.id, context)
        return
      }

      if (event.type === "session.compacted") {
        const context = await achoContext($)
        await inject(client, event.properties.sessionID, context)
        return
      }

      if (event.type === "session.idle") {
        await remember(client, $, event.properties.sessionID)
        return
      }

      if (event.type === "todo.updated" && todoCompleted(event.properties)) {
        await remember(client, $, event.properties.sessionID)
      }
    },
  }
}
