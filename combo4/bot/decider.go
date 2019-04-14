package bot

import (
	"sync"
	"tetris"
	"tetris/combo4"
)

// Decider picks the next best state based on a Scorer.
type Decider struct {
	nfa     *combo4.NFA
	scorers []Scorer
}

// TODO(benjaminchang): Include non-continuous moves here.
var nfa = combo4.NewNFA(combo4.AllContinuousMoves())

// NewDecider creates a new Decider based on the Scorers. The later Scorers
// are used to tie break earlier scorers.
func NewDecider(scorers ...Scorer) *Decider {
	return &Decider{
		scorers: scorers,
	}
}

// NextState returns the "best" possible next state or nil if there are no
// possible moves.
func (d *Decider) NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	choices := nfa.NextStates(initial, current)
	switch len(choices) {
	case 0:
		return nil
	case 1:
		return &choices[0]
	}

	scores := make([]int64, len(choices))
	var bestState combo4.State
	for sIdx, scorer := range d.scorers {
		var wg sync.WaitGroup
		wg.Add(len(choices))
		for idx, choice := range choices {
			idx, choice := idx, choice // Capture range variables.
			go func() {
				scores[idx] = scorer.Score(choice, preview, endBagUsed)
				wg.Done()
			}()
		}
		wg.Wait()

		bestScore := int64(-1 << 63)
		for idx, score := range scores {
			if score > bestScore {
				bestScore = score
				bestState = choices[idx]
			}
		}

		if sIdx != len(d.scorers)-1 {
			break
		}
		// If there are multiple with the best score, narrow the choices
		// for the next Scorer.
		var numBest int
		for _, score := range scores {
			if score == bestScore {
				numBest++
			}
		}
		if numBest > 1 {
			var newIdx int
			for idx, score := range scores {
				if score == bestScore {
					choices[newIdx] = choices[idx]
					newIdx++
				}
			}
			scores = scores[:numBest]
			choices = choices[:numBest]
		}
	}

	return &bestState
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
