package llm

import "context"

// Request is the agent-level description of one LLM turn. It carries no OpenAI
// SDK types so that callers (the agent) never depend on the SDK.
type Request struct {
	Input          string
	Instructions   string
	Tools          []ToolDef
	PrevResponseID string
	Temperature    float64
	// ToolResults, when non-empty, re-feeds tool outputs for a follow-up turn
	// instead of sending fresh user input.
	ToolResults []ToolResult
}

// ToolDef describes a function tool the model may call.
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// ToolResult is the output of an executed tool, tied to the call that produced it.
type ToolResult struct {
	CallID string
	Output string
}

// StreamEvent is a single agent-level event decoded from the SDK stream.
type StreamEvent struct {
	Type         string // "text_delta" | "function_call" | "done" | "error"
	Text         string
	FunctionCall *FunctionCallData
	ResponseID   string
}

// FunctionCallData is a fully-assembled function call emitted once its
// arguments have finished streaming.
type FunctionCallData struct {
	CallID    string
	Name      string
	Arguments string
}

// Client is the boundary the agent depends on. Implementations translate
// between these agent-level types and a concrete LLM provider.
type Client interface {
	Stream(ctx context.Context, req Request) (<-chan StreamEvent, error)
}
