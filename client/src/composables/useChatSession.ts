import { onMounted, reactive } from 'vue'
import { useSessionList } from './useSessionList'
import { useActiveSession } from './useActiveSession'

export function useChatSession() {
  const sessionList = useSessionList()
  const activeSession = useActiveSession(sessionList.sessions)

  onMounted(() => {
    sessionList.load()
  })

  async function createAndSelect(chosenFilePath: string): Promise<void> {
    const session = await sessionList.create(chosenFilePath)
    activeSession.select(session.id)
  }

  return {
    sessionList: reactive(sessionList),
    activeSession: reactive(activeSession),
    createAndSelect,
  }
}
