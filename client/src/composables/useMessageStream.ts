import { ref } from 'vue'
import type { Ref } from 'vue'
import type { Message } from '../types'

const liveMessages: Ref<Message[]> = ref([])
const streaming: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

let currentEventSource: EventSource | null = null

async function sendMessage(sessionId: string, text: string): Promise<void> {
  error.value = null
  streaming.value = true

  try {
    // const response = await fetch(`/api/sessions/${sessionId}/messages`, {
    //   method: 'POST',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify({ text }),
    // })
    // currentEventSource = new EventSource(`/api/sessions/${sessionId}/messages/stream`)
    // currentEventSource.onmessage = (event) => {
    //   const parsed: Message = JSON.parse(event.data)
    //   liveMessages.value.push(parsed)
    // }
    // currentEventSource.onerror = () => {
    //   streaming.value = false
    //   currentEventSource?.close()
    // }
  } catch (err) {
    error.value = err instanceof Error ? err : new Error(String(err))
    streaming.value = false
  }
}

function reset(): void {
  currentEventSource?.close()
  currentEventSource = null
  liveMessages.value = []
  streaming.value = false
  error.value = null
}

export function useMessageStream() {
  return { liveMessages, streaming, error, sendMessage, reset }
}
