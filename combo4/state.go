package combo4

import (
	"tetris"
)

// State represents the state of the playing field. Not including the queue.
type State struct {
	Field Field4x4
	// The piece being held. Possibly the EmptyPiece.
	Hold tetris.Piece
}

// NFA represents a non-determinstic finite automina with some differences.
// All states are considered "final" and there is no "initial" state.
// NFA is safe for concurrent use.
type NFA struct {
	// trans contains possible transitions in the NFA.
	// Usage: trans[piece][state] where piece is the next piece from the queue.
	trans [8]map[State][]State
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

// EndStates returns a set of possible end states given a set of
// initial/current states and pieces to consume. If there are pieces that
// cannot be consumed, EndStates also returns the unconsumed pieces.
func (nfa NFA) EndStates(initial StateSet, pieces ...tetris.Piece) (StateSet, []tetris.Piece) {
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
			return cur, pieces[idx:]
		}
		cur, next = next, cur
		for key := range next {
			delete(next, key)
		}
	}
	return cur, nil
}

// NewNFA creates a new NFA. In general users should reuse the same NFA
// everywhere because the NFA is stateless and immutable.
func NewNFA() *NFA {
	movesList := allContinuousMoves()
	// TODO(benjaminchang): Include non-continuous moves as well.

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

	// Add all the transitions from no Hold piece to a Hold piece.
	// WLOG we can assume that a piece is always held if there isn't one.
	for f := range startFields {
		for _, p := range tetris.NonemptyPieces {
			init := State{f, tetris.EmptyPiece}
			trans[p][init] = append(trans[p][init], State{Field: f, Hold: p})
		}
	}

	// Add all other transitions from states with a Hold piece to
	// other states with a Hold piece.
	for field := range startFields {
		for _, hold := range tetris.NonemptyPieces {
			state := State{field, hold}
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
