import { ref } from 'vue'
import type { Ref } from 'vue'
import type { ChatMessage, TurnInput } from '../types'
import { streamMessage } from '../services/streaming'
import { applyEvent } from '../services/aggregateStream'
import { useToolApproval } from './useToolApproval'

const messages: Ref<ChatMessage[]> = ref([])
const streaming: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

const approval = useToolApproval()

// Every turn runs through here: a plain message and a tool decision differ only
// in the body they send and in the text echoed as the user bubble.
async function run(sessionId: string, echo: string, input: TurnInput): Promise<void> {
  error.value = null
  streaming.value = true

  messages.value.push({ role: 'user', parts: [{ type: 'text', text: echo }] })

  await streamMessage(sessionId, input, {
    onMessage: (event) => {
      applyEvent(messages.value, event)
      // A non-auto-re-feed call ends the agent's loop until the user resolves it.
      if (event.type === 'tool_call' && event.auto_re_feed === false && event.call_id) {
        approval.start(sessionId, event.call_id)
      }
    },
    onClose: (ok) => {
      streaming.value = false
      if (!ok) {
        error.value = new Error('Stream ended unexpectedly')
      }
    },
  })
}

function sendMessage(sessionId: string, text: string): Promise<void> {
  return run(sessionId, text, { input: { message: text } })
}

// Resolves the pending call: Approve, or Request changes with the user's message.
function resolveTool(approved: boolean, message: string): Promise<void> {
  const call = approval.take()
  if (!call) return Promise.resolve()

  return run(call.sessionId, message, {
    tool_result: { call_id: call.callId, result: { approved, message } },
  })
}

function reset(): void {
  messages.value = []
  streaming.value = false
  error.value = null
  approval.clear()
}

export function useMessageStream() {
  return { messages, streaming, error, sendMessage, resolveTool, reset }
}
