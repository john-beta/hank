import { ref, computed, reactive } from 'vue'
import type { Ref } from 'vue'
import type { Session } from '../types'
import { createSession } from '../services/api'

const activeSession: Ref<Session | null> = ref(null)

const isWizardMode = computed(() => activeSession.value === null)

function startNew(): void {
  activeSession.value = null
}

async function create(chosenRootDir: string): Promise<void> {
  activeSession.value = await createSession(chosenRootDir)
}

export function useChatSession() {
  return reactive({ activeSession, isWizardMode, startNew, create })
}
