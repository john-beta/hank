<script setup lang="ts">
import { useMessageStream } from '../../composables/useMessageStream'
import { PromptInput, PromptInputSubmit, PromptInputTextarea } from '../ai-elements/prompt-input'
import type { PromptInputMessage } from '../ai-elements/prompt-input'

const props = defineProps<{
  sessionId: string
}>()

const { streaming, sendMessage } = useMessageStream()

function handleSubmit(message: PromptInputMessage): void {
  if (message.text.trim() === '') return
  sendMessage(props.sessionId, message.text)
}
</script>

<template>
  <footer class="chat-footer">
    <PromptInput class="chat-footer-form" @submit="handleSubmit">
      <PromptInputTextarea placeholder="Type a message..." :disabled="streaming" />
      <PromptInputSubmit :status="streaming ? 'streaming' : 'ready'" :disabled="streaming" />
    </PromptInput>
  </footer>
</template>
