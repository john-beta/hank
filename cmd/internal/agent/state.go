package agent

import (
	"github.com/john-beta/hank/cmd/internal/store"
)

// State carries per-turn conversation state, built fresh from the store on
// every request and never held on the Agent. ApprovedProposal is never
// persisted — see prepareToolResultRequest for where it flips.
type State struct {
	SessionID        string
	RootDir          string
	ApprovedProposal bool
	PrevResponseID   string
}

// StateFromStore always starts ApprovedProposal false — the store carries no
// such column. A nil lastAgentTurn means this is the session's first turn.
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
