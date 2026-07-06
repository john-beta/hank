package agent

// Phase selects which instructions and tools the agent runs with.
type Phase int

const (
	// PhaseOne runs with the tool available.
	PhaseOne Phase = iota
	// PhaseTwo runs without tools.
	PhaseTwo
)

// State carries conversation state across turns.
type State struct {
	Phase          Phase
	PrevResponseID string
}

// NewState returns a fresh state starting in PhaseOne.
func NewState() *State {
	return &State{Phase: PhaseOne}
}

// Transition advances the phase. Placeholder for business logic (Task 2).
func (s *State) Transition() {}
