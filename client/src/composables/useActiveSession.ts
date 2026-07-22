import { ref, computed } from 'vue'
import type { Ref } from 'vue'
import type { Session } from '../types'

const activeSessionId: Ref<string | null> = ref(null)

const isWizardMode = computed(() => activeSessionId.value === null)

function startNew(): void {
  activeSessionId.value = null
}

function select(sessionId: string): void {
  activeSessionId.value = sessionId
}

export function useActiveSession(sessions: Ref<Session[]>) {
  const activeSession = computed(() => {
    return sessions.value.find((session) => session.session_id === activeSessionId.value) ?? null
  })

  return { activeSessionId, activeSession, isWizardMode, startNew, select }
}
