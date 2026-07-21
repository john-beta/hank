<script setup lang="ts">
import type { ErrorPart } from '../../../types'
import { AlertCircleIcon } from '@lucide/vue'

defineProps<{
  part: ErrorPart
}>()
</script>

<template>
  <div class="rounded-lg border border-red-200 bg-red-50 p-4">
    <div class="flex items-start gap-3">
      <AlertCircleIcon class="mt-0.5 mr-2 shrink-0 text-red-600" :size="20" />
      <div class="min-w-0 flex-1">
        <h3 class="font-semibold text-red-900">Error</h3>
        <p>{{ part.error }}</p>
      </div>
    </div>
  </div>
</template>
