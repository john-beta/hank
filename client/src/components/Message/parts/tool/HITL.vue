<script setup lang="ts">
import type { ToolPart } from '../../../../types'
import { computed } from 'vue'
import { Button } from '../../../ui/button'
import { useMessageStream } from '../../../../composables/useMessageStream'
import { useToolApproval } from '../../../../composables/useToolApproval'

const props = defineProps<{
  part: ToolPart
}>()

const { resolveTool } = useMessageStream()
const { pending, unlock } = useToolApproval()

// Only the call still awaiting a decision is actionable; earlier proposals in
// the transcript are already resolved.
const active = computed(() => pending.value?.callId === props.part.callId)
</script>

<template>
  <div v-if="active" class="flex gap-2 pt-3">
    <Button size="sm" @click="resolveTool(true, 'Plan approved')">Approve</Button>
    <Button size="sm" variant="outline" @click="unlock()">Request changes</Button>
  </div>
</template>
