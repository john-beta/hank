package agent

import (
	"github.com/john-beta/hank/cmd/internal/store"
)

// State is per-request conversation state, built fresh from the store each
// request. ApprovedProposal is never persisted — see prepareToolResultRequest.
type State struct {
	SessionID        string
	RootDir          string
	ApprovedProposal bool
	PrevResponseID   string
}

// StateFromStore starts ApprovedProposal false (the store has no such column).
// A nil lastAgentTurn means the session's first turn.
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
