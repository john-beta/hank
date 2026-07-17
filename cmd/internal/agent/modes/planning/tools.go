package planning

import "fmt"

// ProposeStructure is not auto-re-fed: the client resolves it, and an approving
// result flips the session into Executing. RunExecution is absent by design —
// its absence is the mode boundary.
var autoReFeed = map[string]bool{
	"RunExploration": true,
	"ProposeStructure": false,
}

// execute routes Planning's tool names; bodies are stubs until the real logic lands.
func execute(name, args string) (string, error) {
	switch name {
	case "RunExploration":
		return "not yet implemented", nil
	case "ProposeStructure":
		return "not yet implemented", nil
	default:
		return "", fmt.Errorf("planning: unknown tool: %s", name)
	}
}
