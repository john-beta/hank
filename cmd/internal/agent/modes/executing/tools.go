package executing

import (
	"context"
	"encoding/json"

	"github.com/john-beta/hank/cmd/internal/agent/modes/mode"
)

// RunExecution's presence here and absence from Planning is the mode boundary.
var autoReFeed = map[string]bool{
	"RunExploration": true,
	"RunExecution":   true,
}

// execute routes Executing's tool names; bodies are stubs until the real logic lands.
func execute(ctx context.Context, name string, args string) (string, error) {
	switch name {
	case "RunExploration":
		return mode.RunPythonProcess(ctx, args)
	case "RunExecution":
		return mode.RunPythonProcess(ctx, args)
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
