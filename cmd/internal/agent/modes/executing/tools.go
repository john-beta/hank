package executing

import "fmt"

// RunExecution's presence here and absence from Planning is the mode boundary.
var autoReFeed = map[string]bool{
	"RunExploration": true,
	"RunExecution":   true,
}

// execute routes Executing's tool names; bodies are stubs until the real logic lands.
func execute(name, args string) (string, error) {
	switch name {
	case "RunExploration":
		return "not yet implemented", nil
	case "RunExecution":
		return "not yet implemented", nil
	default:
		return "", fmt.Errorf("executing: unknown tool: %s", name)
	}
}
