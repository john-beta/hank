package agent

import (
	"github.com/john-beta/hank/cmd/internal/agent/phases/phase"
	"github.com/john-beta/hank/cmd/internal/store"
)

// State carries conversation state for a single turn. It is built fresh from the
// store at the start of each request and never held on the Agent.
type State struct {
	SessionID      string
	Phase          phase.ID
	PrevResponseID string

	// input is the user's text for this turn. It is transient (not persisted
	// state) and seeds the first Request in runLoop.
	input string

	// turnRecords is the ordered audit trail of every model turn produced in
	// this request's loop — the intermediate tool-call turns and the final
	// text turn — paired with the phase active when each was produced.
	// Transient: runLoop appends to it, run persists it.
	turnRecords []turnRecord
}

// turnRecord pairs a completed model response with the phase active when it
// was produced, so persistence reflects the phase at the time, not the
// loop's final phase.
type turnRecord struct {
	ResponseID string
	Phase      phase.ID
}

// recordResponse registers a completed model turn: it updates PrevResponseID
// (used to chain the next request) and appends to the audit trail so every
// turn — not just the last — can be persisted after the loop.
func (s *State) recordResponse(id string, ph phase.ID) {
	s.PrevResponseID = id
	s.turnRecords = append(s.turnRecords, turnRecord{ResponseID: id, Phase: ph})
}

// StateFromStore builds the per-turn State from a persisted session and its last
// agent turn (if any). A nil lastAgentTurn means this is the first turn.
func StateFromStore(s store.Session, lastAgentTurn *store.Turn) *State {
	st := &State{
		SessionID: s.SessionID,
		Phase:     phase.ID(s.Phase),
	}
	if lastAgentTurn != nil && lastAgentTurn.ResponseID != nil {
		st.PrevResponseID = *lastAgentTurn.ResponseID
	}
	return st
}
