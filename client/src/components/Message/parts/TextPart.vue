<script setup lang="ts">
import type { TextPart } from '../../../types'
import { MessageResponse } from '../../ai-elements/message'

defineProps<{
  part: TextPart
}>()
</script>

<template>
  <MessageResponse :content="part.text" />
</template>
