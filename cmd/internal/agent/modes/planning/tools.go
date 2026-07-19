package planning

import (
	"context"
	"encoding/json"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
)

// ProposeStructure is not auto-re-fed: the client resolves it, and an approving
// result flips the session into Executing. RunExecution is absent by design —
// its absence is the mode boundary.
var autoReFeed = map[string]bool{
	"RunExploration":   true,
	"ProposeStructure": false,
}

// execute routes Planning's tool names; bodies are stubs until the real logic lands.
func execute(ctx context.Context, name string, args string) (string, error) {
	switch name {
	case "RunExploration":
		return mode.RunPythonProcess(ctx, args)
	case "ProposeStructure":
		return unreachableToolErrorToJSON(&unreachableToolError{Success: false, Error: "Propose structure could not be approved"}), nil
	default:
		return unreachableToolErrorToJSON(&unreachableToolError{Success: false, Error: "Tool result not found for this call"}), nil
	}
}

type unreachableToolError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func unreachableToolErrorToJSON(err *unreachableToolError) string {
	b, _ := json.Marshal(err)
	return string(b)
}
