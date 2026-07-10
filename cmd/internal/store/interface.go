package store

import (
	"context"
	"time"
)

// Session is a conversation thread. Phase tracks its current position in the
// agent's phase progression.
type Session struct {
	SessionID string
	Phase     int
	CreatedAt time.Time
}

// Turn is a single exchange within a session. A user turn carries Content and a
// nil ResponseID; an agent turn carries a ResponseID and nil Content (its text
// is reconstructed from OpenAI via ResponseID later).
type Turn struct {
	TurnID     string
	SessionID  string
	ResponseID *string // nil for user turns
	Phase      int
	Role       string  // "user" or "agent"
	Content    *string // nil for agent turns
	CreatedAt  time.Time
}

// Store is the persistence boundary for sessions and turns.
type Store interface {
	CreateSession(ctx context.Context) (Session, error)
	GetSession(ctx context.Context, sessionID string) (Session, error)

	SaveTurn(ctx context.Context, t Turn) error
	LastAgentTurn(ctx context.Context, sessionID string) (*Turn, error)
}
