import type { Session, Message } from '../types'

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL || '/api'

export async function fetchSessions(): Promise<Session[]> {
  const response = await fetch(`${BASE_URL}/sessions`)
  if (!response.ok) {
    throw new Error(`Failed to fetch sessions: ${response.status}`)
  }
  const data: { sessions: Session[] } = await response.json()
  return data.sessions
}

export async function createSession(path: string): Promise<Session> {
  const response = await fetch(`${BASE_URL}/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path }),
  })
  if (!response.ok) {
    throw new Error(`Failed to create session: ${response.status}`)
  }
  const data: Session = await response.json()
  return data
}

export async function fetchSessionHistory(sessionId: string): Promise<Message[]> {
  const response = await fetch(`${BASE_URL}/sessions/${sessionId}/messages`)
  if (!response.ok) {
    throw new Error(`Failed to fetch session history: ${response.status}`)
  }
  const data: { messages: Message[] } = await response.json()
  return data.messages
}
