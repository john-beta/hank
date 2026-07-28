package modes

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
)

// Prompt config lives in the OpenAI dashboard; Go only references it by id + version.
const (
	planningPromptID      = "pmpt_6a5a753b523881939f22b420fc63fee6074a8fbf8db7fc2d"
	planningPromptVersion = "7"
)

// ProposeStructure is not auto-re-fed: the client resolves it, and an approving
// result flips the session into Executing. RunImplementation is absent by design —
// its absence is the mode boundary.
var planningAutoReFeed = map[string]bool{
	"RunExploration":   true,
	"ProposeStructure": false,
}

func NewPlanning() Mode {
	return Mode{
		Prompt:     llm.Prompt{ID: planningPromptID, Version: planningPromptVersion},
		Execute:    planningExecute,
		AutoReFeed: func(name string) bool { return planningAutoReFeed[name] },
	}
}

// planningExecute routes Planning's tool names to their handlers. ProposeStructure
// is client-resolved, so reaching its case here means a desynced call.
func planningExecute(ctx context.Context, name string, args string, state ExecState) (string, error) {
	switch name {
	case "RunExploration":
		return RunPythonProcess(ctx, args, state.RootDir)
	case "ProposeStructure":
		return unreachableToolErrorToJSON(&unreachableToolError{Success: false, Error: "Propose structure could not be approved"}), nil
	default:
		return unreachableToolErrorToJSON(&unreachableToolError{Success: false, Error: "Tool result not found for this call"}), nil
	}
}
