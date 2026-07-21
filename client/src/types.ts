export interface Session {
  session_id: string
  root_dir: string
  created_at: string
}

export type EventType = 'text_delta' | 'tool_call' | 'tool_result' | 'done' | 'error'

export type RoleType = 'user' | 'agent'

export interface Message {
  type: EventType
  role: RoleType
  text?: string
  tool_name?: string
  tool_args?: string
  tool_result?: string
  call_id?: string
  auto_re_feed?: boolean
  error?: string
  response_id?: string
}
