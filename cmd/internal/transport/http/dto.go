package transport

// SendMessageRequest is the JSON body of POST /api/messages.
type SendMessageRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// CreateSessionResponse is the JSON body returned by POST /api/sessions.
type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}
