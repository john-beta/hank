package store

import (
	"context"
	"time"
)

// Session is a conversation thread bound to a workspace. The mode
// (Planning/Executing) is not part of it — it's computed per request, never persisted.
type Session struct {
	SessionID string
	RootDir   string
	CreatedAt time.Time
}

// Turn is a single exchange within a session. OutputText is the user's message
// or the assistant's streamed text, nil for a tool-result user turn (whose payload
// lives in call.result). See migrations.go.
type Turn struct {
	TurnID     string
	SessionID  string
	Role       string  // "user" or "assistant"
	OutputText *string // nil for tool-result user turns
	ResponseID *string // nil for user turns
	CreatedAt  time.Time
}

// Call is one tool call: exactly one row per CallID, inserted with Result nil
// and updated in place when the tool resolves.
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
	LastAssistantTurn(ctx context.Context, sessionID string) (*Turn, error)

	// IsPendingCall reports whether callID names an unresolved, non-auto call —
	// one still awaiting a client-supplied result.
	IsPendingCall(ctx context.Context, callID string) (bool, error)
	SaveCall(ctx context.Context, c Call) error
	UpdateCallResult(ctx context.Context, callID string, result string) error
	PendingProposeStructureByCallID(ctx context.Context, callID string) (string, error)
}
