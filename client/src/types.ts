export interface Session {
  session_id: string
  root_dir: string
  created_at: string
}

export type EventType = 'text_delta' | 'tool_call' | 'tool_result' | 'done' | 'error'

export type RoleType = 'user' | 'assistant'

// Wire protocol: one NDJSON event streamed from the backend.
export interface StreamEvent {
  type: EventType
  role?: RoleType
  text?: string
  tool_name?: string
  tool_args?: string
  tool_result?: string
  call_id?: string
  auto_re_feed?: boolean
  error?: string
  response_id?: string
}

// UI model: one chat turn, made of ordered parts.
export interface TextPart {
  type: 'text'
  text: string
}

export interface ToolPart {
  type: 'tool'
  callId?: string
  name?: string
  args?: string
  result?: string
  autoReFeed?: boolean
  state: 'call' | 'result'
}

export type Part = TextPart | ToolPart

export interface ChatMessage {
  role: RoleType
  parts: Part[]
}
