<script setup lang="ts">
import { useChatSession } from './composables/useChatSession'
import Sidebar from './components/Sidebar.vue'
import MainPanel from './components/MainPanel.vue'

const { sessionList, activeSession, createAndSelect } = useChatSession()
</script>

<template>
  <div class="app">
    <Sidebar
      :sessions="sessionList.sessions"
      :loading="sessionList.loading"
      :error="sessionList.error"
      :active-session-id="activeSession.activeSessionId"
      @new-chat="activeSession.startNew"
      @select-session="activeSession.select"
    />
    <MainPanel
      :is-wizard-mode="activeSession.isWizardMode"
      :active-session="activeSession.activeSession"
      @create-session="createAndSelect"
    />
  </div>
</template>
