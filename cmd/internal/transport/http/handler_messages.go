package transport

import (
	"encoding/json"
	"net/http"

	"github.com/john-beta/hank/cmd/internal/agent"
)

// Passing r.Context() to the agent is what wires client disconnects to
// loop/stream cancellation: when the client goes away the context cancels,
// the agent loop and LLM stream unwind, and the events channel closes,
// ending this handler.
func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	input, err := toTurnInput(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	setSSEHeaders(w)

	events := h.agent.Handle(r.Context(), req.SessionID, input)
	for event := range events {
		writeSSE(w, event)
	}
}

// toTurnInput is translation only — no business logic, no mode awareness.
func toTurnInput(req SendMessageRequest) (agent.TurnInput, error) {
	switch {
	case req.ToolResult != nil:
		result := string(req.ToolResult.Result)
		return agent.TurnInput{ToolResult: &agent.ToolResultInput{
			CallID: req.ToolResult.CallID,
			Result: result,
		}}, nil
	case req.Input != nil:
		msg := req.Input.Message
		return agent.TurnInput{Message: &msg}, nil
	default:
		return agent.TurnInput{}, errMissingInput
	}
}
