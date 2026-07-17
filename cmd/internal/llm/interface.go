package llm

import "context"

// Request carries no OpenAI SDK types, so callers (the agent) never depend on
// the SDK directly.
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

// ToolDef describes a function tool the model may call. AutoReFeed is
// agent-domain metadata, not part of the OpenAI tool definition — it's
// stripped before the request reaches the SDK (see buildParams).
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

// FunctionCallData is a fully-assembled function call, emitted once its
// arguments have finished streaming.
type FunctionCallData struct {
	CallID    string
	Name      string
	Arguments string
}

// Client is the boundary the agent depends on.
type Client interface {
	Stream(ctx context.Context, req Request) (<-chan StreamEvent, error)
}
