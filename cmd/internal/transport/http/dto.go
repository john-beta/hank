package transport

import (
	"encoding/json"
	"errors"
)

// errMissingInput is returned when a message request carries neither an input
// message nor a tool result.
var errMissingInput = errors.New("request must contain either input.message or tool_result")

// SendMessageRequest is the JSON body of POST /api/messages. It carries exactly
// one of two shapes: `input.message` for a plain user message, or `tool_result`
// for the client resolving a pending non-auto call.
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

// CreateSessionRequest is the JSON body of POST /api/sessions. It names the
// workspace the new session curates.
type CreateSessionRequest struct {
	RootDir string `json:"root_dir"`
}

// CreateSessionResponse is the JSON body returned by POST /api/sessions.
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}
