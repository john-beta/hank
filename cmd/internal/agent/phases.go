package agent

import "github.com/john-beta/hank/cmd/internal/llm"

// PhaseKit is the instructions + tools bundle for a phase.
type PhaseKit struct {
	Instructions string
	Tools        []llm.ToolDef
}

// KitFor returns the kit configured for the given phase.
func KitFor(p Phase) PhaseKit {
	switch p {
	case PhaseTwo:
		return PhaseKit{
			Instructions: "You are a helpful assistant. Respond concisely.",
			Tools:        nil,
		}
	default:
		return PhaseKit{
			Instructions: "You are a helpful assistant. Use the get_weather tool when asked about weather.",
			Tools: []llm.ToolDef{
				{
					Name:        "get_weather",
					Description: "Get the current weather for a location.",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]any{
								"type":        "string",
								"description": "The city or place to get the weather for.",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		}
	}
}

// NextPhase decides the phase for the next iteration given the current phase
// and the tool calls just executed. Scaffold: never transitions. Task 2 adds
// the real rule here.
func NextPhase(current Phase, calls []llm.FunctionCallData, results []llm.ToolResult) Phase {
	return current
}
