import type { ChatMessage, StreamEvent, ToolPart } from '../types'

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
        autoReFeed: event.auto_re_feed,
        state: 'call',
      })
      break
    case 'tool_result': {
      const parts = currentAssistant(messages).parts
      const part = parts.find((p) => p.type === 'tool' && p.callId === event.call_id) as
        | ToolPart
        | undefined
      if (part) {
        part.result = event.tool_result
        part.state = 'result'
      }
      break
    }
    // Reasoning arrives as start → delta* → done, one block per reasoning item.
    // The part is created by the first delta, so items with no summary text render
    // nothing; `isReasoning` doubles as the "block is still open" marker.
    case 'reasoning_delta': {
      const msg = currentAssistant(messages)
      const last = msg.parts.at(-1)
      if (last?.type === 'reasoning' && last.isReasoning) last.text += event.text ?? ''
      else msg.parts.push({ type: 'reasoning', isReasoning: true, text: event.text ?? '' })
      break
    }
    case 'reasoning_start':
    case 'reasoning_done': {
      const last = currentAssistant(messages).parts.at(-1)
      if (last?.type === 'reasoning') last.isReasoning = false
      break
    }
    case 'error':
      currentAssistant(messages).parts.push({
        type: 'error',
        error: event.error ?? 'Unknown error',
      })
      break
    // 'done' | future types: ignored for now.
    // To handle one, add a case here (and a Part type in types.ts if it renders).
  }
}
