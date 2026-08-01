export interface Session {
  session_id: string
  root_dir: string
  created_at: string
}

export type EventType =
  | 'text_delta'
  | 'tool_call'
  | 'tool_result'
  | 'reasoning_start'
  | 'reasoning_delta'
  | 'reasoning_done'
  | 'done'
  | 'error'

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

// The client's resolution of a tool call the agent left to a human.
export interface ToolDecision {
  approved: boolean
  message: string
}

// Wire protocol: POST /api/messages body — exactly one of input | tool_result.
export type TurnInput =
  | { input: { message: string } }
  | { tool_result: { call_id: string; result: ToolDecision } }

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

export interface ReasoningPart {
  type: 'reasoning'
  isReasoning: boolean
  text: string
}

export interface ErrorPart {
  type: 'error'
  error: string
}

export type Part = TextPart | ToolPart | ReasoningPart | ErrorPart

export interface ChatMessage {
  role: RoleType
  parts: Part[]
}
