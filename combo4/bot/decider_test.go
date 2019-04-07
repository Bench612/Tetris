package bot

import (
	"math/rand"
	"testing"
	"tetris"
	"tetris/combo4"
)

func BenchmarkNextState(b *testing.B) {
	stateSet := combo4.NewNFA(combo4.AllContinuousMoves()).States()
	states := make([]combo4.State, 0, len(stateSet))
	for s := range stateSet {
		states = append(states, s)
	}

	d := NewDecider()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		randState := states[rand.Intn(len(states))]
		queue := tetris.RandPieces(7)
		var bag tetris.PieceSet

		d.NextState(randState, queue[0], queue[1:], bag)
	}
}

func TestStartGameRatio(t *testing.T) {
	const (
		trials         = 100
		piecesPerTrial = 100
	)

	rand.Seed(1)
	d := NewDecider()

	var incomplete int
	for t := 0; t < trials; t++ {
		queue := tetris.RandPieces(piecesPerTrial)
		input := make(chan tetris.Piece, 1)
		output := d.StartGame(combo4.LeftI, queue[0], queue[1:7], input)
		for _, p := range queue[7:] {
			input <- p
			if <-output == nil {
				incomplete++
				break
			}
		}
	}
	if ratio := 1 - float64(incomplete)/trials; ratio < 0.7 {
		t.Errorf("Decider has win rate=%.2f, want at least 0.7", ratio)
	}
}
