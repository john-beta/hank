<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import Reasoning from '@/components/ai-elements/reasoning/Reasoning.vue'
import type { ReasoningPart } from '../../../types'
import ReasoningTrigger from '@/components/ai-elements/reasoning/ReasoningTrigger.vue'
import ReasoningContent from '@/components/ai-elements/reasoning/ReasoningContent.vue'

const props = defineProps<{
  part: ReasoningPart
}>()

const rootRef: Ref<HTMLElement | null> = ref(null)

// While reasoning is streaming, follow the growing text — unlike ChatBody's
// "stick if already near bottom", this keeps this block itself in view, so it
// doesn't matter how far a single delta grows it.
function followIfStreaming() {
  if (props.part.isReasoning) {
    rootRef.value?.scrollIntoView({ block: 'end', inline: 'nearest', behavior: 'smooth' })
  }
}

watch(
  rootRef,
  (el, _prev, onCleanup) => {
    if (!el) return
    const observer = new ResizeObserver(followIfStreaming)
    observer.observe(el)
    onCleanup(() => observer.disconnect())
  },
  { flush: 'post' },
)
</script>

<template>
  <div ref="rootRef" class="w-full">
    <Reasoning class="w-full" :is-streaming="part.isReasoning">
      <ReasoningTrigger />
      <ReasoningContent :content="part.text" class="[&_*]:text-gray-400 dark:[&_*]:text-gray-400" />
    </Reasoning>
  </div>
</template>
