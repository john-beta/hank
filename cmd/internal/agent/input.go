package agent

// TurnInput is the discriminated input to a turn. The client sends exactly one
// of the two shapes: a plain user Message (scenario 1) or a ToolResult that
// resolves a pending non-auto call (scenario 2). Exactly one field is non-nil.
type TurnInput struct {
	Message    *string
	ToolResult *ToolResultInput
}

// ToolResultInput carries the client's resolution of a pending call. Result is
// the tool output as a JSON string.
type ToolResultInput struct {
	CallID string
	Result string
}
