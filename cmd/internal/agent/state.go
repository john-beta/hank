package agent

import (
	"github.com/john-beta/hank/cmd/internal/store"
)

// State carries conversation state for a single turn. It is built fresh from the
// store at the start of each request and never held on the Agent. ApprovedProposal
// is not persisted anywhere — it starts false every request and is only flipped
// in-memory, within prepareToolResultRequest, when that same turn's tool result
// approves. The operating mode is derived from it by mode.Resolve once, before
// the loop.
type State struct {
	SessionID        string
	RootDir          string
	ApprovedProposal bool
	PrevResponseID   string
}

// StateFromStore builds the per-turn State from a persisted session and its last
// agent turn (if any). A nil lastAgentTurn means this is the first turn.
// ApprovedProposal always starts false here — the store carries no such signal.
func StateFromStore(s store.Session, lastAgentTurn *store.Turn) *State {
	st := &State{
		SessionID: s.SessionID,
		RootDir:   s.RootDir,
	}
	if lastAgentTurn != nil && lastAgentTurn.ResponseID != nil {
		st.PrevResponseID = *lastAgentTurn.ResponseID
	}
	return st
}
