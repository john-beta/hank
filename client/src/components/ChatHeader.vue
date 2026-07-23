<script setup lang="ts">
import type { Session } from '../types'
import { FolderIcon } from '@lucide/vue'

defineProps<{
  session: Session
}>()
</script>

<template>
  <header class="flex items-center justify-between gap-4 border-b bg-background px-4 py-3">
    <div class="flex min-w-0 items-center gap-1.5">
      <FolderIcon class="size-4 shrink-0 text-blue-500" />
      <p class="truncate font-mono text-sm text-foreground">{{ session.root_dir }}</p>
    </div>
    <p class="shrink-0 text-xs text-muted-foreground">
      {{ new Date(session.created_at).toLocaleString() }}
    </p>
  </header>
</template>
