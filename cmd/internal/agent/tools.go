package agent

import "fmt"

// ToolFunc executes a tool given its raw JSON argument string and returns a
// JSON result string.
type ToolFunc func(args string) (string, error)

// ToolRegistry maps tool names to their implementations.
type ToolRegistry struct {
	tools map[string]ToolFunc
}

// NewToolRegistry returns a registry pre-loaded with the mock get_weather tool.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: map[string]ToolFunc{
			"get_weather": func(args string) (string, error) {
				return `{"temp_c": 24, "condition": "sunny"}`, nil
			},
		},
	}
}

// Execute runs the named tool. It errors if the tool is not registered.
func (r *ToolRegistry) Execute(name, args string) (string, error) {
	fn, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return fn(args)
}
