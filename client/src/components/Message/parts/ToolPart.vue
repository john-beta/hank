<script setup lang="ts">
import type { ToolPart } from '../../../types'
import AutoRefeed from './tool/AutoRefeed.vue'
import NoAutoRefeed from './tool/NoAutoRefeed.vue'

defineProps<{
  part: ToolPart
}>()
</script>

<template>
  <AutoRefeed v-if="part.autoReFeed" :part="part" />
  <NoAutoRefeed v-else :part="part" />
</template>
