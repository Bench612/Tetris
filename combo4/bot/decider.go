package bot

import (
	"tetris"
	"tetris/combo4"
)

// Decider picks the next best state based on a Scorer.
type Decider struct {
	nfa    *combo4.NFA
	scorer Scorer
}

// NewDecider creates a new Decider.
func NewDecider(scorer Scorer) *Decider {
	return &Decider{
		// TODO(benjaminchang): Include non-continuous moves here.
		nfa:    combo4.NewNFA(combo4.AllContinuousMoves()),
		scorer: scorer,
	}
}

type scoreTuple struct {
	// The state that this score pertains to.
	state combo4.State

	// Components of the score orderd by importance.
	consumedPieces int
	score          int
	numStates      int
}

func (s scoreTuple) GreaterThan(other scoreTuple) bool {
	if s.consumedPieces != other.consumedPieces {
		return s.consumedPieces > other.consumedPieces
	}
	if s.score != other.score {
		return s.score > other.score
	}
	return s.numStates > other.numStates
}

// NextState returns the "best" possible next state or nil if there are no
// possible moves.
func (d *Decider) NextState(initial combo4.State, current tetris.Piece, next []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	choices := d.nfa.NextStates(initial, current)
	if len(choices) == 0 {
		return nil
	}

	scores := make(chan scoreTuple, len(choices))
	for _, choice := range choices {
		choice := choice // Capture range variable.
		go func() {
			endStates, consumed := d.nfa.EndStates(combo4.NewStateSet(choice), next)
			if consumed == len(next) {
				scores <- scoreTuple{
					state:          choice,
					consumedPieces: consumed,
					score:          d.scorer.Score(endStates, endBagUsed),
					numStates:      len(endStates),
				}
				return
			}
			scores <- scoreTuple{
				state:          choice,
				consumedPieces: consumed,
			}
		}()
	}

	var best scoreTuple
	for _ = range choices {
		if pair := <-scores; pair.GreaterThan(best) {
			best = pair
		}
	}

	return &best.state
}

// StartGame returns a channel that outputs the next state and then an
// additional state for each input. The channel returns nil if there are no
// more possible moves.
//
// StartGame assumes there is no piece held and the game is starting with no
// pieces played yet (starting with an empty bag).
//
// StartGame panics if a piece that does not follow the 7 bag randomizer is
// added to the input channel.
func (d *Decider) StartGame(initial combo4.Field4x4, current tetris.Piece, next []tetris.Piece, input chan tetris.Piece) chan *combo4.State {
	cpy := make([]tetris.Piece, len(next))
	copy(cpy, next)
	next = cpy

	state := &combo4.State{Field: initial}
	bag := current.PieceSet()
	for _, n := range next {
		bag = bag.Add(n)
		if bag.Len() == 7 {
			bag = 0
		}
	}

	output := make(chan *combo4.State, len(input))
	go func() {
		defer close(output)

		// Output the first move.
		state := d.NextState(*state, current, next, bag)
		output <- state

		for p := range input {
			if state == nil {
				output <- nil
				continue
			}

			// Shift next and the current piece.
			if len(next) == 0 {
				current = p
			} else {
				current = next[0]

				copy(next, next[1:])
				next[len(next)-1] = p
			}

			// Update the bag.
			if bag.Len() == 7 {
				bag = 0
			}
			if bag.Contains(p) {
				panic(`impossible piece "` + p.String() + `" for bag state ` + bag.String())
			}
			bag = bag.Add(p)

			state = d.NextState(*state, current, next, bag)
			output <- state
		}
	}()

	return output
}
