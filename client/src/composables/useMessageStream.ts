import { ref } from 'vue'
import type { Ref } from 'vue'
import type { ChatMessage } from '../types'
import { streamMessage } from '../services/streaming'
import { applyEvent } from '../services/aggregateStream'

const messages: Ref<ChatMessage[]> = ref([])
const streaming: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

async function sendMessage(sessionId: string, text: string): Promise<void> {
  error.value = null
  streaming.value = true

  messages.value.push({ role: 'user', parts: [{ type: 'text', text }] })

  await streamMessage(sessionId, text, {
    onMessage: (event) => {
      applyEvent(messages.value, event)
    },
    onClose: (ok) => {
      streaming.value = false
      if (!ok) {
        error.value = new Error('Stream ended unexpectedly')
      }
    },
  })
}

function reset(): void {
  messages.value = []
  streaming.value = false
  error.value = null
}

export function useMessageStream() {
  return { messages, streaming, error, sendMessage, reset }
}
