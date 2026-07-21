<script setup lang="ts">
import type { ToolPart } from '../../../types'

defineProps<{
  part: ToolPart
}>()
</script>

<template>
  <div class="tool-part">
    <div class="tool-part-name">{{ part.name }}</div>
    <pre v-if="part.args" class="tool-part-args">{{ part.args }}</pre>
    <pre v-if="part.state === 'result'" class="tool-part-result">{{ part.result }}</pre>
  </div>
</template>
