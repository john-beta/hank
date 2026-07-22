import type { StreamEvent, TurnInput } from '../types'

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL

export async function streamMessage(
  sessionId: string,
  input: TurnInput,
  callbacks: {
    onMessage: (event: StreamEvent) => void
    onClose: (ok: boolean) => void
  },
): Promise<void> {
  const response = await fetch(`${BASE_URL}/api/messages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ...input, session_id: sessionId }),
  })

  if (!response.ok || !response.body) {
    callbacks.onClose(false)
    return
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let sawDone = false

  function handleLine(line: string): void {
    if (line === '') return
    let parsed: StreamEvent
    try {
      parsed = JSON.parse(line)
    } catch {
      return
    }
    if (parsed.type === 'done') sawDone = true
    callbacks.onMessage(parsed)
  }

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''
      for (const line of lines) {
        handleLine(line)
      }
    }
    handleLine(buffer)
    callbacks.onClose(sawDone)
  } catch {
    callbacks.onClose(false)
  }
}
