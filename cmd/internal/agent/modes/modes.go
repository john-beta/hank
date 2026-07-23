// Package modes defines the two ReAct behaviors (Planning, Executing) that
// drive an agent turn: each mode pairs an OpenAI prompt with the Go-side
// tool routing for that phase.
package modes

import (
	"context"

	"github.com/john-beta/hank/cmd/internal/llm"
)

// ExecState is the per-request data a tool execution needs — a deliberately
// small subset of agent.State (this package cannot import agent, which
// imports modes).
type ExecState struct {
	SessionID string
	RootDir   string
}

// Mode bundles the OpenAI prompt that configures a turn with the Go-side
// behavior. NewPlanning() / NewExecuting() each build one.
type Mode struct {
	Prompt  llm.Prompt
	Execute func(ctx context.Context, name string, args string, state ExecState) (string, error)
	// AutoReFeed reports whether the named tool executes-and-re-feeds
	// automatically. Each mode builds it by closing over its own tool set.
	AutoReFeed func(name string) bool
}
