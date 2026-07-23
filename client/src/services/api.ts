import type { Session } from '../types'

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL

export async function createSession(path: string): Promise<Session> {
  const response = await fetch(`${BASE_URL}/api/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ root_dir: path }),
  })
  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Failed to create session: ${response.status}`)
  }
  const data: Session = await response.json()
  return data
}
