import { onMounted, reactive } from 'vue'
import { useSessionList } from './useSessionList'
import { useActiveSession } from './useActiveSession'

export function useChatSession() {
  const sessionList = useSessionList()
  const activeSession = useActiveSession(sessionList.sessions)

  onMounted(() => {
    sessionList.load()
  })

  async function createAndSelect(chosenRootDir: string): Promise<void> {
    const session = await sessionList.create(chosenRootDir)
    activeSession.select(session.session_id)
  }

  return {
    sessionList: reactive(sessionList),
    activeSession: reactive(activeSession),
    createAndSelect,
  }
}
