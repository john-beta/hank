package transport

import (
	"encoding/json"
	"errors"
)

var errMissingInput = errors.New("request must contain either input.message or tool_result")

// SendMessageRequest is the JSON body of POST /api/messages. It carries
// exactly one of two shapes: `input.message` for a plain user message, or
// `tool_result` for the client resolving a pending non-auto call.
type SendMessageRequest struct {
	SessionID string `json:"session_id"`
	Input     *struct {
		Message string `json:"message"`
	} `json:"input"`
	ToolResult *struct {
		CallID string          `json:"call_id"`
		Result json.RawMessage `json:"result"`
	} `json:"tool_result"`
}

type CreateSessionRequest struct {
	RootDir string `json:"root_dir"`
}

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}
