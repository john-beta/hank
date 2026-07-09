package phase_one

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
)

var tools = []llm.ToolDef{
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
}

func execute(name, args string) (string, error) {
	switch name {
	case "get_weather":
		return `{"temp_c": 24, "condition": "sunny"}`, nil
	default:
		return "", fmt.Errorf("phase_one: unknown tool: %s", name)
	}
}
