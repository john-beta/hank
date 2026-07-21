import type { ChatMessage, StreamEvent, ToolPart } from '../types'

// Adaptation layer: folds the flat NDJSON event stream into the turn-centric
// UI model. A single assistant response is a sequence of events; here it becomes
// one assistant ChatMessage with ordered parts (text, tool, text, ...).

function currentAssistant(messages: ChatMessage[]): ChatMessage {
  const last = messages.at(-1)
  if (last?.role === 'assistant') return last
  messages.push({ role: 'assistant', parts: [] })
  // Re-read to return the reactive proxy when `messages` is a ref array, so
  // subsequent in-place mutations are tracked.
  return messages.at(-1)!
}

export function applyEvent(messages: ChatMessage[], event: StreamEvent): void {
  switch (event.type) {
    case 'text_delta': {
      const msg = currentAssistant(messages)
      const last = msg.parts.at(-1)
      if (last?.type === 'text') last.text += event.text ?? ''
      else msg.parts.push({ type: 'text', text: event.text ?? '' })
      break
    }
    case 'tool_call':
      currentAssistant(messages).parts.push({
        type: 'tool',
        callId: event.call_id,
        name: event.tool_name,
        args: event.tool_args,
        state: 'call',
      })
      break
    case 'tool_result': {
      const parts = currentAssistant(messages).parts
      const part = parts.find(
        (p) => p.type === 'tool' && p.callId === event.call_id,
      ) as ToolPart | undefined
      if (part) {
        part.result = event.tool_result
        part.state = 'result'
      }
      break
    }
    // 'done' | 'error' | future types: ignored for now.
    // To handle one, add a case here (and a Part type in types.ts if it renders).
  }
}
