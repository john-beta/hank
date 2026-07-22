<script setup lang="ts">
import type { ChatMessage } from '../../types'
import { Message, MessageContent } from '../ai-elements/message'
import PartRenderer from './PartRenderer.vue'

defineProps<{
  message: ChatMessage
}>()
</script>

<template>
  <Message :from="message.role">
    <MessageContent>
      <PartRenderer v-for="(part, i) in message.parts" :key="i" :part="part" />
    </MessageContent>
  </Message>
</template>
