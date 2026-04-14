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
      }
    },
  }
}
