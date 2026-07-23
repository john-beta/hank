<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import MessageRenderer from '../Message/MessageRenderer.vue'
import Shimmer from './Shimmer.vue'
import { useMessageStream } from '../../composables/useMessageStream'

// How close to the bottom (in px) the user must be for growing content to
// auto-scroll them along. Scrolled up further than this is treated as
// intentionally reading back, so we leave their position alone.
const SCROLL_BOTTOM_OFFSET = 100

const props = defineProps<{
  sessionId: string
}>()

const { messages, reset: resetStream, streaming } = useMessageStream()

const bodyRef: Ref<HTMLElement | null> = ref(null)
const messagesRef: Ref<HTMLElement | null> = ref(null)

function stickToBottomIfNear() {
  const container = bodyRef.value
  if (!container) return
  // Nothing to scroll at all — no-op.
  if (container.scrollHeight <= container.clientHeight) return

  const distanceFromBottom = container.scrollHeight - container.scrollTop - container.clientHeight
  if (distanceFromBottom <= SCROLL_BOTTOM_OFFSET) {
    container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' })
  }
}

watch(
  messagesRef,
  (el, _prev, onCleanup) => {
    if (!el) return
    const observer = new ResizeObserver(stickToBottomIfNear)
    observer.observe(el)
    onCleanup(() => observer.disconnect())
  },
  { flush: 'post' },
)

watch(
  () => props.sessionId,
  () => {
    resetStream()
  },
  { immediate: true },
)
</script>

<template>
  <section ref="bodyRef" class="chat-body">
    <p v-if="messages.length === 0" class="chat-body-empty">No messages yet</p>
    <div v-else ref="messagesRef" class="chat-body-messages">
      <MessageRenderer v-for="(message, index) in messages" :key="index" :message="message" />
      <Shimmer v-if="streaming" />
    </div>
  </section>
</template>
