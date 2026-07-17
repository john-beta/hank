package store

import (
	"context"
	"time"
)

// Session is a conversation thread bound to a workspace. The operating mode
// (Planning/Executing) is not part of it — that's computed per request in the
// agent layer, never persisted here.
type Session struct {
	SessionID string
	RootDir   string
	CreatedAt time.Time
}

// Turn is a single exchange within a session. See migrations.go for
// output_text semantics: a user's message or an agent's streamed text, nil
// for a user turn that carries a tool result (whose payload lands via
// call.result instead).
type Turn struct {
	TurnID     string
	SessionID  string
	Role       string  // "user" or "agent"
	OutputText *string // nil for tool-result user turns
	ResponseID *string // nil for user turns
	CreatedAt  time.Time
}

// Call is one tool call emitted by the agent. Exactly one row exists per
// CallID: inserted with Result nil, updated in place when the tool resolves.
type Call struct {
	CallID     string
	TurnID     string
	Name       string
	Args       *string
	Result     *string
	AutoReFeed bool
}

type Store interface {
	CreateSession(ctx context.Context, rootDir string) (Session, error)
	GetSession(ctx context.Context, sessionID string) (Session, error)

	SaveTurn(ctx context.Context, t Turn) (turnID string, err error)
	LastAgentTurn(ctx context.Context, sessionID string) (*Turn, error)

	SaveCall(ctx context.Context, c Call) error
	UpdateCallResult(ctx context.Context, callID string, result string) error
	// PendingCall returns the single unresolved, non-auto call for the
	// session's latest agent turn, or nil, nil if there is none.
	PendingCall(ctx context.Context, sessionID string) (*Call, error)
}
