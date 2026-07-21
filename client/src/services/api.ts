import type { Session, StreamEvent, ChatMessage } from '../types'
import { applyEvent } from './aggregateStream'

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL

export async function fetchSessions(): Promise<Session[]> {
  const response = await fetch(`${BASE_URL}/sessions`)
  if (!response.ok) {
    throw new Error(`Failed to fetch sessions: ${response.status}`)
  }
  const data: { sessions: Session[] } = await response.json()
  return data.sessions
}

export async function createSession(path: string): Promise<Session> {
  const response = await fetch(`${BASE_URL}/api/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ root_dir: path }),
  })
  if (!response.ok) {
    throw new Error(`Failed to create session: ${response.status}`)
  }
  const data: Session = await response.json()
  return data
}

export async function fetchSessionHistory(sessionId: string): Promise<ChatMessage[]> {
  const response = await fetch(`${BASE_URL}/sessions/${sessionId}/messages`)
  if (!response.ok) {
    throw new Error(`Failed to fetch session history: ${response.status}`)
  }
  // Provisional: the GET history endpoint is not implemented yet. When it lands,
  // it will return the persisted event log, which we fold into the same
  // turn-centric ChatMessage[] the live stream produces so rendering is identical.
  const data: { messages: StreamEvent[] } = await response.json()
  const messages: ChatMessage[] = []
  for (const event of data.messages) {
    if (event.role === 'user') {
      messages.push({ role: 'user', parts: [{ type: 'text', text: event.text ?? '' }] })
    } else {
      applyEvent(messages, event)
    }
  }
  return messages
}
