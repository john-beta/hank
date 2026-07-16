<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import type { Message } from '../../types'
import { fetchSessionHistory } from '../../services/api'
import MessageRenderer from '../messages/MessageRenderer.vue'

const props = defineProps<{
  sessionId: string
}>()

const messages: Ref<Message[]> = ref([])
const loading: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

async function loadHistory(sessionId: string): Promise<void> {
  loading.value = true
  error.value = null
  try {
    messages.value = await fetchSessionHistory(sessionId)
  } catch (err) {
    error.value = err instanceof Error ? err : new Error(String(err))
  } finally {
    loading.value = false
  }
}

watch(() => props.sessionId, loadHistory, { immediate: true })
</script>

<template>
  <section class="chat-body">
    <p v-if="loading" class="chat-body-empty">Loading...</p>
    <p v-else-if="error" class="chat-body-empty">Failed to load</p>
    <p v-else-if="messages.length === 0" class="chat-body-empty">No messages yet</p>
    <div v-else class="chat-body-messages">
      <MessageRenderer v-for="message in messages" :key="message.id" :message="message" />
    </div>
  </section>
</template>
