<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { Ref } from 'vue'
import type { Message } from '../../types'
import { fetchSessionHistory } from '../../services/api'
import MessageRenderer from '../messages/MessageRenderer.vue'
import { useMessageStream } from '../../composables/useMessageStream'

const props = defineProps<{
  sessionId: string
}>()

const messages: Ref<Message[]> = ref([])
const loading: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

const { liveMessages, reset: resetStream } = useMessageStream()

const allMessages = computed(() => [...messages.value, ...liveMessages.value])

async function loadHistory(sessionId: string): Promise<void> {
  loading.value = true
  error.value = null
  try {
    messages.value = await fetchSessionHistory(sessionId)
  } catch (err) {
    // error.value = err instanceof Error ? err : new Error(String(err))
  } finally {
    loading.value = false
  }
}

watch(
  () => props.sessionId,
  (sessionId) => {
    resetStream()
    loadHistory(sessionId)
  },
  { immediate: true },
)
</script>

<template>
  <section class="chat-body">
    <p v-if="loading" class="chat-body-empty">Loading...</p>
    <p v-else-if="error" class="chat-body-empty">Failed to load</p>
    <p v-else-if="allMessages.length === 0" class="chat-body-empty">No messages yet</p>
    <div v-else class="chat-body-messages">
      <MessageRenderer v-for="(message, index) in allMessages" :key="index" :message="message" />
    </div>
  </section>
</template>
