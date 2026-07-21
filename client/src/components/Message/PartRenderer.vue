<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import type { Part } from '../../types'
import TextPart from './parts/TextPart.vue'
import ToolPart from './parts/ToolPart.vue'

// To add a part type:
// 1. Add the Part variant in types.ts
// 2. Create the component in parts/ (e.g. ReasoningPart.vue)
// 3. Import it here and add one entry below
const COMPONENT_BY_TYPE: Record<Part['type'], Component> = {
  text: TextPart,
  tool: ToolPart,
}

const props = defineProps<{
  part: Part
}>()

const component = computed(() => COMPONENT_BY_TYPE[props.part.type] ?? null)
</script>

<template>
  <component :is="component" v-if="component" :part="part" />
</template>
