package bot

import (
	"sync"
	"tetris"
	"tetris/combo4"
)

// Policy determines the next state.
type Policy interface {
	NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State
}

// scorePolicy picks the next best state based on a Scorer.
type scorePolicy struct {
	nfa    *combo4.NFA
	scorer Scorer
}

// PolicyFromScorer creates a new Policy based a Scorer.
func PolicyFromScorer(nfa *combo4.NFA, scorer Scorer) Policy {
	return &scorePolicy{
		nfa:    nfa,
		scorer: scorer,
	}
}

// NextState returns the "best" possible next state or nil if there are no
// possible moves.
func (p *scorePolicy) NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	choices := p.nfa.NextStates(initial, current)
	switch len(choices) {
	case 0:
		return nil
	case 1:
		return &choices[0]
	}

	scores := make([]int64, len(choices))
	var wg sync.WaitGroup
	wg.Add(len(choices))
	for idx, choice := range choices {
		idx, choice := idx, choice // Capture range variables.
		go func() {
			scores[idx] = p.scorer.Score(choice, preview, endBagUsed)
			wg.Done()
		}()
	}
	wg.Wait()

	var bestState combo4.State
	bestScore := int64(-1 << 63)
	for idx, score := range scores {
		if score > bestScore {
			bestScore = score
			bestState = choices[idx]
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
func StartGame(pol Policy, initial combo4.Field4x4, current tetris.Piece, next []tetris.Piece, input chan tetris.Piece) chan *combo4.State {
	bag := current.PieceSet()
	for _, n := range next {
		bag = bag.Add(n)
		if bag.Len() == 7 {
			bag = 0
		}
	}
	return ResumeGame(pol, combo4.State{Field: initial}, current, next, bag, input)
}

// ResumeGame is like StartGame but does not assume the game is played from
// the beginning.
func ResumeGame(pol Policy, initialState combo4.State, current tetris.Piece, next []tetris.Piece, endBagUsed tetris.PieceSet, input chan tetris.Piece) chan *combo4.State {
	// Make a copy of next because we will be modifying it.
	cpy := make([]tetris.Piece, len(next))
	copy(cpy, next)
	next = cpy

	output := make(chan *combo4.State, len(input))
	go func() {
		defer close(output)

		// Output the first move.
		state := pol.NextState(initialState, current, next, endBagUsed)
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
			if endBagUsed.Len() == 7 {
				endBagUsed = 0
			}
			if endBagUsed.Contains(p) {
				panic(`impossible piece "` + p.String() + `" for bag state ` + endBagUsed.String())
			}
			endBagUsed = endBagUsed.Add(p)

			state = pol.NextState(*state, current, next, endBagUsed)
			output <- state
		}
	}()

	return output
}
