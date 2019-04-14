package combo4

import (
	"fmt"
	"tetris"
)

// State represents the state of the playing field. Not including the queue.
type State struct {
	Field Field4x4
	// The piece being held. Possibly the EmptyPiece.
	Hold tetris.Piece
	// Whether the hold piece can be swapped or not.
	SwapRestricted bool
}

func (s State) String() string {
	return fmt.Sprintf("Hold:\n%s\nField:\n%s", s.Hold.GameString(), s.Field)
}

// StateSet represents a set of States.
type StateSet map[State]bool

// NewStateSet creates a StateSet from a list of States.
func NewStateSet(states ...State) StateSet {
	set := make(StateSet)
	for _, state := range states {
		set[state] = true
	}
	return set
}

// Equals returns true if two StateSets are equal.
func (s StateSet) Equals(other StateSet) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

// Slice converts the StateSet to a slice.
func (s StateSet) Slice() []State {
	slice := make([]State, 0, len(s))
	for state := range s {
		slice = append(slice, state)
	}
	return slice
}

// NFA represents a non-determinstic finite automina with some differences.
// All states are considered "final" and there is no "initial" state.
// NFA is safe for concurrent use.
type NFA struct {
	// trans contains possible transitions in the NFA.
	// Usage: trans[piece][state] where piece is the next piece from the queue.
	trans [8]map[State][]State
}

// NextStates returns the possible next states.
func (nfa *NFA) NextStates(initial State, piece tetris.Piece) []State {
	ns := nfa.trans[piece][initial]
	cpy := make([]State, len(ns))
	copy(cpy, ns)
	return cpy
}

// States returns the set of States represented in the NFA.
func (nfa *NFA) States() StateSet {
	states := make(map[State]bool)
	for _, m := range nfa.trans {
		for input, outputs := range m {
			states[input] = true
			for _, output := range outputs {
				states[output] = true
			}
		}
	}
	return states
}

// EndStates returns a set of end states given a set of initial/current
// states and pieces to consume. EndStates also returns the number of consumed
// pieces. The final state is returned if not all pieces were consumed.
func (nfa *NFA) EndStates(initial StateSet, pieces []tetris.Piece) (StateSet, int) {
	cur := make(map[State]bool)
	for state, ok := range initial {
		cur[state] = ok
	}

	next := make(map[State]bool)
	for idx, piece := range pieces {
		trans := nfa.trans[piece]
		for curState := range cur {
			for _, nextState := range trans[curState] {
				next[nextState] = true
			}
		}
		if len(next) == 0 {
			return cur, idx
		}
		cur, next = next, cur
		for key := range next {
			delete(next, key)
		}
	}
	return cur, len(pieces)
}

// NewNFA creates a new NFA. In general callers should reuse the same NFA
// everywhere because the NFA is safe for concurrent use.
func NewNFA(movesList []Move) *NFA {
	// Get a set of all Field4x4s which have possible moves.
	startFields := make(map[Field4x4]bool)
	for _, move := range movesList {
		startFields[move.Start] = true
	}

	// Group the moves by Start and Piece.
	moves := make(map[Field4x4]map[tetris.Piece][]Field4x4)
	for field := range startFields {
		moves[field] = make(map[tetris.Piece][]Field4x4)
	}
	for _, m := range movesList {
		moves[m.Start][m.Piece] = append(moves[m.Start][m.Piece], m.End)
	}

	var trans [8]map[State][]State
	for _, piece := range tetris.NonemptyPieces {
		trans[int(piece)] = make(map[State][]State)
	}

	// Add all the transitions from no Hold piece.
	for field := range startFields {
		for _, piece := range tetris.NonemptyPieces {
			endStates := make([]State, 0, len(moves[field][piece])+1)
			// Add transition from holding the piece.
			endStates = append(endStates, State{Field: field, Hold: piece, SwapRestricted: true})
			// Add transitions from playing the piece.
			for _, endField := range moves[field][piece] {
				endStates = append(endStates, State{Field: endField})
			}

			state := State{Field: field}
			trans[piece][state] = append(trans[piece][state], endStates...)
		}
	}

	// Add all transitions from a SwapRestricted state.
	for field := range startFields {
		for _, hold := range tetris.NonemptyPieces {
			state := State{Field: field, Hold: hold, SwapRestricted: true}
			for _, piece := range tetris.NonemptyPieces {
				endStates := make([]State, 0, len(moves[field][piece]))
				// Add transitions from playing a piece.
				for _, endField := range moves[field][piece] {
					// The state is no longer SwapRestricted.
					endStates = append(endStates, State{Field: endField, Hold: hold})
				}
				trans[piece][state] = append(trans[piece][state], endStates...)
			}
		}
	}

	// Add all other transitions from states with a swappable Hold piece to
	// other states with a Hold piece.
	for field := range startFields {
		for _, hold := range tetris.NonemptyPieces {
			state := State{Field: field, Hold: hold}
			for _, piece := range tetris.NonemptyPieces {
				endStates := make([]State, 0, len(moves[field][piece])+len(moves[field][hold]))
				// Add all transitions that keep the Hold piece.
				for _, endField := range moves[field][piece] {
					endStates = append(endStates, State{Field: endField, Hold: hold})
				}
				// Add all transitions that swap the Hold piece and play it.
				for _, endField := range moves[field][hold] {
					endStates = append(endStates, State{Field: endField, Hold: piece})
				}
				trans[piece][state] = append(trans[piece][state], endStates...)
			}
		}
	}

	return &NFA{trans: trans}
}
