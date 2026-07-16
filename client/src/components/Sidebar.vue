<script setup lang="ts">
import type { Session } from '../types'

defineProps<{
  sessions: Session[]
  loading: boolean
  error: Error | null
  activeSessionId: string | null
}>()

defineEmits<{
  'new-chat': []
  'select-session': [sessionId: string]
}>()
</script>

<template>
  <aside class="sidebar">
    <button class="new-chat-btn" @click="$emit('new-chat')">New chat</button>
    <nav class="chat-list">
      <p v-if="loading">Loading...</p>
      <p v-else-if="error">Failed to load</p>
      <template v-else>
        <p v-if="sessions.length === 0">No sessions yet</p>
        <button
          v-for="session in sessions"
          :key="session.id"
          class="chat-list-item"
          :class="{ 'chat-list-item--active': session.id === activeSessionId }"
          @click="$emit('select-session', session.id)"
        >
          {{ session.root_dir }}
        </button>
      </template>
    </nav>
  </aside>
</template>
