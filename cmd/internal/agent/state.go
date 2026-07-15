package agent

import (
	"github.com/john-beta/hank/cmd/internal/store"
)

// State carries conversation state for a single turn. It is built fresh from the
// store at the start of each request and never held on the Agent. The operating
// mode is NOT stored here — it is derived from ApprovedProposal by mode.Resolve
// once, before the loop.
type State struct {
	SessionID        string
	RootDir          string
	ApprovedProposal bool
	PrevResponseID   string
}

// StateFromStore builds the per-turn State from a persisted session and its last
// agent turn (if any). A nil lastAgentTurn means this is the first turn.
func StateFromStore(s store.Session, lastAgentTurn *store.Turn) *State {
	st := &State{
		SessionID:        s.SessionID,
		RootDir:          s.RootDir,
		ApprovedProposal: s.ApprovedProposal,
	}
	if lastAgentTurn != nil && lastAgentTurn.ResponseID != nil {
		st.PrevResponseID = *lastAgentTurn.ResponseID
	}
	return st
}
