<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'

defineProps<{
  sessionId: string
}>()

const draft: Ref<string> = ref('')

function handleSubmit(): void {
  draft.value = ''
}
</script>

<template>
  <footer class="chat-footer">
    <form class="chat-footer-form" @submit.prevent="handleSubmit">
      <input class="chat-footer-input" v-model="draft" placeholder="Type a message..." />
      <button class="chat-footer-send-btn">Send</button>
    </form>
  </footer>
</template>
