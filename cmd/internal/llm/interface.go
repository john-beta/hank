package llm

import "context"

// Request carries no OpenAI SDK types, so the agent never depends on the SDK.
type Request struct {
	Input          string
	Instructions   string
	Tools          []ToolDef
	PrevResponseID string
	// ToolResults, when non-empty, re-feeds tool outputs for a follow-up turn
	// instead of sending fresh user input.
	ToolResults []ToolResult
}

// ToolDef describes a function tool. AutoReFeed is agent metadata, not part of
// the OpenAI tool definition — it's stripped before reaching the SDK (mapTools).
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]any
	AutoReFeed  bool
}

type ToolResult struct {
	CallID string
	Output string
}

type StreamEvent struct {
	Type         string // "text_delta" | "function_call" | "done" | "error"
	Text         string
	FunctionCall *FunctionCallData
	ResponseID   string
}

// FunctionCallData is a function call assembled once its arguments finish streaming.
type FunctionCallData struct {
	CallID    string
	Name      string
	Arguments string
}

// Client is the boundary the agent depends on.
type Client interface {
	Stream(ctx context.Context, req Request) (<-chan StreamEvent, error)
}
