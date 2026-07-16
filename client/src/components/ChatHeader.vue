<script setup lang="ts">
import type { Session } from '../types'

defineProps<{
  session: Session
}>()
</script>

<template>
  <header class="chat-header">
    <p>{{ session.root_dir }}</p>
    <p>{{ new Date(session.created_at).toLocaleString() }}</p>
  </header>
</template>
