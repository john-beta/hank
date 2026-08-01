<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import type { Part } from '../../types'
import TextPart from './parts/TextPart.vue'
import ToolPart from './parts/ToolPart.vue'
import ErrorPart from './parts/ErrorPart.vue'
import ReasoningPart from './parts/ReasoningPart.vue'

const COMPONENT_BY_TYPE: Record<Part['type'], Component> = {
  text: TextPart,
  tool: ToolPart,
  error: ErrorPart,
  reasoning: ReasoningPart,
}

const props = defineProps<{
  part: Part
}>()

const component = computed(() => COMPONENT_BY_TYPE[props.part.type] ?? null)
</script>

<template>
  <component :is="component" v-if="component" :part="part" />
</template>
