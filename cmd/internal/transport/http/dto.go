package transport

// SendMessageRequest is the JSON body of POST /api/messages.
type SendMessageRequest struct {
	Message string `json:"message"`
}
