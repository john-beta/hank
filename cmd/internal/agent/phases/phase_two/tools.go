package phase_two

import (
	"fmt"

	"github.com/john-beta/hank/cmd/internal/llm"
)

var tools []llm.ToolDef = nil

func execute(name, args string) (string, error) {
	return "", fmt.Errorf("phase_two: no tools available")
}
