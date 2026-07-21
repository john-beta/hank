<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'
import { useMessageStream } from '../../composables/useMessageStream'

const props = defineProps<{
  sessionId: string
}>()

const draft: Ref<string> = ref('')

const { streaming, sendMessage } = useMessageStream()

function handleSubmit(): void {
  if (draft.value.trim() === '') return
  sendMessage(props.sessionId, draft.value)
  draft.value = ''
}
</script>

<template>
  <footer class="chat-footer">
    <form class="chat-footer-form" @submit.prevent="handleSubmit">
      <input class="chat-footer-input" v-model="draft" placeholder="Type a message..." :disabled="streaming" />
      <button class="chat-footer-send-btn" :disabled="streaming">Send</button>
    </form>
  </footer>
</template>
