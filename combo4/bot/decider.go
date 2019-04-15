package bot

import (
	"sync"
	"tetris"
	"tetris/combo4"
)

// Decider picks the next best state.
type Decider interface {
	NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State
}

// scoreDecider picks the next best state based on a Scorer.
type scoreDecider struct {
	nfa    *combo4.NFA
	scorer Scorer
}

// NewScoreDecider creates a new Decider based a Scorer.
func NewScoreDecider(nfa *combo4.NFA, scorer Scorer) Decider {
	return &scoreDecider{
		nfa:    nfa,
		scorer: scorer,
	}
}

// NextState returns the "best" possible next state or nil if there are no
// possible moves.
func (d *scoreDecider) NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	choices := d.nfa.NextStates(initial, current)
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
			scores[idx] = d.scorer.Score(choice, preview, endBagUsed)
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
func StartGame(d Decider, initial combo4.Field4x4, current tetris.Piece, next []tetris.Piece, input chan tetris.Piece) chan *combo4.State {
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

type mdpDecider struct {
	policy         map[GameState]combo4.State
	defaultDecider Decider
}

// NewMDPDecider returns a new Decider based on an MDP.
func NewMDPDecider(mdp *MDP) Decider {
	policy, scorer := mdp.Policy()
	return &mdpDecider{
		policy:         policy,
		defaultDecider: NewScoreDecider(combo4.NewNFA(combo4.AllContinuousMoves()), scorer),
	}
}

func (d *mdpDecider) NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	gameState := GameState{
		State:   initial,
		Current: current,
		Preview: tetris.MustSeq(preview),
		BagUsed: endBagUsed,
	}
	if nextState, ok := d.policy[gameState]; ok {
		copy := nextState
		return &copy
	}
	return d.defaultDecider.NextState(initial, current, preview, endBagUsed)
}
