package bot

import (
	"errors"
	"math"
	"tetris"
	"tetris/combo4"
)

// Decider picks the next best state based on a Scorer.
type Decider struct {
	nfa    *combo4.NFA
	scorer *Scorer
}

// NewDecider creates a new Decider.
func NewDecider() *Decider {
	return &Decider{
		nfa:    combo4.NewNFA(combo4.AllContinuousMoves()),
		scorer: NewScorer(),
	}
}

type scoreTuple struct {
	score     int32
	numStates int
	state     combo4.State
}

// NextState returns the best possible next state or an error if
// there is no possible next state.
func (d *Decider) NextState(initial combo4.State, queue []tetris.Piece, endBagUsed tetris.PieceSet) (combo4.State, error) {
	if len(queue) == 0 {
		return combo4.State{}, errors.New("queue must have at least one element")
	}

	possible := d.nfa.NextStates(initial, queue[0])
	if len(possible) == 0 {
		return combo4.State{}, errors.New("no possible moves")
	}
	remainingQueue := queue[1:]

	scores := make(chan scoreTuple, len(possible))
	for _, possibility := range possible {
		possibility := possibility // Capture range variable.
		go func() {
			endStates := d.nfa.EndStates(possibility, remainingQueue)
			scores <- scoreTuple{
				score:     d.scorer.Score(endStates, endBagUsed),
				numStates: len(endStates),
				state:     possibility,
			}
		}()
	}

	best := scoreTuple{score: math.MinInt32}
	for _ = range possible {
		pair := <-scores
		if pair.score > best.score || (pair.score == best.score && pair.numStates > best.numStates) {
			best = pair
		}
	}
	return best.state, nil
}

// StartGame returns a channel that outputs the next state and then an
// additional state for each input. The channel returns nil if there are no
// more possible moves.
//
// StartGame assumes there is no piece held and the game is starting with no
// pieces played yet (starting with an empty bag).
func (d *Decider) StartGame(initial combo4.Field4x4, first tetris.Piece, next []tetris.Piece, input chan tetris.Piece) chan *combo4.State {
	state := &combo4.State{Field: initial}
	queue := append([]tetris.Piece{first}, next...)
	bag := tetris.NewPieceSet(queue...)

	output := make(chan *combo4.State, len(input))
	defer close(output)

	getNext := func() *combo4.State {
		if state == nil {
			return nil
		}
		next, err := d.NextState(*state, queue, bag)
		if err != nil {
			return nil
		}
		return &next
	}

	go func() {
		// Output the first move.
		next := getNext()
		output <- next
		state = next

		for p := range input {
			if bag.Len() == 7 {
				bag = p.PieceSet()
			} else {
				bag = bag.Add(p)
			}

			queue = append([]tetris.Piece{p}, queue[1:]...)

			next := getNext()
			output <- next
			state = next
		}
	}()

	return output
}
