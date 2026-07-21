<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'
import type { Message } from '../../types'

const props = defineProps<{
  message: Message
}>()

// To add a new message type:
// 1. Create the component in this folder (e.g. TextMessage.vue)
// 2. Import it here
// 3. Add one entry: text: TextMessage
const COMPONENT_BY_TYPE: Record<string, Component> = {}

const component = computed(() => COMPONENT_BY_TYPE[props.message.type] ?? null)
</script>

<template>
  <!-- <component :is="component" v-if="component" :message="message" />
  <div v-else>Unsupported message type: {{ message.type }}</div> -->
  <div>
    <h3>Message</h3>
    {{ JSON.stringify(message) }}
  </div>
</template>
