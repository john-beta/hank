<script setup lang="ts">
import type { Session } from '../types'
import ChatWizard from './ChatWizard.vue'
import ChatHeader from './ChatHeader.vue'
import ChatBody from './ChatBody/ChatBody.vue'
import ChatFooter from './ChatFooter/ChatFooter.vue'

defineProps<{
  isWizardMode: boolean
  activeSession: Session | null
}>()

defineEmits<{
  'create-session': [chosenRootDir: string]
}>()
</script>

<template>
  <main class="main">
    <ChatWizard v-if="isWizardMode" @create-session="$emit('create-session', $event)" />
    <template v-else-if="activeSession">
      <ChatHeader :session="activeSession" />
      <ChatBody :session-id="activeSession.session_id" />
      <ChatFooter :session-id="activeSession.session_id" />
    </template>
  </main>
</template>
