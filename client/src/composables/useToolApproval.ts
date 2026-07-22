import { computed, ref } from 'vue'
import type { Ref } from 'vue'

// A tool_call with auto_re_feed=false stops the agent loop: it waits for a
// human decision. While one is pending the footer is locked, so the only way
// forward is Approve or Request changes — the latter unlocks the footer so the
// user can type the message that goes back as the rejection.
interface PendingCall {
  sessionId: string
  callId: string
}

const pending: Ref<PendingCall | null> = ref(null)
const unlocked: Ref<boolean> = ref(false)

const inputBlocked = computed(() => pending.value !== null && !unlocked.value)

function start(sessionId: string, callId: string): void {
  pending.value = { sessionId, callId }
  unlocked.value = false
}

// "Request changes": let the user write; the next submit resolves the call.
function unlock(): void {
  unlocked.value = true
}

// Consumes the pending call — a decision is being sent for it.
function take(): PendingCall | null {
  const call = pending.value
  clear()
  return call
}

function clear(): void {
  pending.value = null
  unlocked.value = false
}

export function useToolApproval() {
  return { pending, inputBlocked, start, unlock, take, clear }
}
