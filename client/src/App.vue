<script setup lang="ts">
import { useChatSession } from './composables/useChatSession'
import MainPanel from './components/MainPanel.vue'
import { Button } from '@/components/ui/button'

const chatSession = useChatSession()
</script>

<template>
  <div class="app">
    <Button variant="outline" size="sm" class="new-chat-btn" @click="chatSession.startNew">
      New chat
    </Button>
    <MainPanel
      :is-wizard-mode="chatSession.isWizardMode"
      :active-session="chatSession.activeSession"
    />
  </div>
</template>
