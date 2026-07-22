import { ref } from 'vue'
import type { Ref } from 'vue'
import type { Session } from '../types'
import { fetchSessions, createSession } from '../services/api'

const sessions: Ref<Session[]> = ref([])
const loading: Ref<boolean> = ref(false)
const error: Ref<Error | null> = ref(null)

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    sessions.value = await fetchSessions()
  } catch (err) {
    error.value = err instanceof Error ? err : new Error(String(err))
  } finally {
    loading.value = false
  }
}

async function create(chosenRootDir: string): Promise<Session> {
  const session = await createSession(chosenRootDir)
  sessions.value.push(session)
  return session
}

export function useSessionList() {
  return { sessions, loading, error, load, create }
}
