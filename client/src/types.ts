export interface Session {
  id: string
  root_dir: string
  created_at: string
}

export interface Message {
  id: string
  type: string
  [key: string]: unknown
}
