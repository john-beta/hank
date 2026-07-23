<script setup lang="ts">
import { computed } from 'vue'
import { useMessageStream } from '../../composables/useMessageStream'
import { useToolApproval } from '../../composables/useToolApproval'
import { PromptInput, PromptInputSubmit, PromptInputTextarea } from '../ai-elements/prompt-input'
import type { PromptInputMessage } from '../ai-elements/prompt-input'

const props = defineProps<{
  sessionId: string
}>()

const { streaming, sendMessage, resolveTool } = useMessageStream()
const { pending, inputBlocked } = useToolApproval()

const disabled = computed(() => streaming.value || inputBlocked.value)

const placeholder = computed(() =>
  inputBlocked.value ? 'Approve or request changes to continue...' : 'Type a message...',
)

// Reaching submit with a call still pending means the user chose "Request
// changes", so the text resolves that call instead of starting a new turn.
function handleSubmit(message: PromptInputMessage): void {
  const text = message.text.trim()
  if (text === '') return
  if (pending.value) resolveTool(false, text)
  else sendMessage(props.sessionId, text)
}
</script>

<template>
  <footer class="chat-footer">
    <PromptInput class="chat-footer-form" @submit="handleSubmit">
      <PromptInputTextarea :placeholder="placeholder" :disabled="disabled" />
      <PromptInputSubmit :status="streaming ? 'streaming' : 'ready'" :disabled="disabled" />
    </PromptInput>
  </footer>
</template>
