package modes

import "encoding/json"

// unreachableToolError is the payload returned for a tool name a mode
// doesn't recognize, or that can't resolve automatically.
type unreachableToolError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func unreachableToolErrorToJSON(err *unreachableToolError) string {
	b, _ := json.Marshal(err)
	return string(b)
}
