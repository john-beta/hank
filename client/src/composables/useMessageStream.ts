import { ref } from 'vue'
import type { Ref } from 'vue'
import type { Message } from '../types'
import { streamMessage } from '../services/streaming'

const liveMessages: Ref<Message[]> = ref([])
const streaming: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

async function sendMessage(sessionId: string, text: string): Promise<void> {
  error.value = null
  streaming.value = true

  liveMessages.value.push({ type: 'text_delta', role: 'user', text })

  await streamMessage(sessionId, text, {
    // TODO: text_delta events currently arrive as separate entries; merging them into a single growing message is intentionally left for later.
    onMessage: (event) => {
      liveMessages.value.push({ ...event, role: 'agent' })
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
  liveMessages.value = []
  streaming.value = false
  error.value = null
}

export function useMessageStream() {
  return { liveMessages, streaming, error, sendMessage, reset }
}
