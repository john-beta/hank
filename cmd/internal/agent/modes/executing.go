package modes

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
)

const (
	executingPromptID      = "pmpt_6a5d1005286c81978fc2110e4abe55ff0184ca20821201e2"
	executingPromptVersion = "7"
)

// RunImplementation's presence here and absence from Planning is the mode boundary.
var executingAutoReFeed = map[string]bool{
	"GetWorkspaceState": true,
	"RunImplementation": true,
}

func NewExecuting() Mode {
	return Mode{
		Prompt:     llm.Prompt{ID: executingPromptID, Version: executingPromptVersion},
		Execute:    executingExecute,
		AutoReFeed: func(name string) bool { return executingAutoReFeed[name] },
	}
}

// executingExecute routes Executing's tool names to their handlers.
func executingExecute(ctx context.Context, name string, args string, state ExecState) (string, error) {
	switch name {
	case "GetWorkspaceState":
		return RunPythonProcess(ctx, args, state.RootDir)
	case "RunImplementation":
		return RunPythonProcess(ctx, args, state.RootDir)
	default:
		return unreachableToolErrorToJSON(&unreachableToolError{Success: false, Error: "Tool result not found for this call"}), nil
	}
}
