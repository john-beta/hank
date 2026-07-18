package mode

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
)

// Mode bundles the OpenAI prompt that configures a turn with the Go-side
// behavior.
// planning.New() / executing.New() each build one.
type Mode struct {
	Prompt  llm.Prompt
	Execute func(ctx context.Context, name string, args string) (string, error)
	// AutoReFeed reports whether the named tool executes-and-re-feeds
	// automatically. Each mode builds it by closing over its own tool set.
	AutoReFeed func(name string) bool
}
