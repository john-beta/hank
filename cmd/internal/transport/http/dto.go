package transport

import (
	"encoding/json"
	"errors"
	"time"
)

var errMissingInput = errors.New("request must contain either input.message or tool_result")

// SendMessageRequest is the POST /api/messages body: exactly one of
// `input.message` (plain user message) or `tool_result` (client resolving a call).
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
	SessionID string    `json:"session_id"`
	RootDir   string    `json:"root_dir"`
	CreatedAt time.Time `json:"created_at"`
}
