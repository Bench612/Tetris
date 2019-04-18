package bot

import (
	"math/rand"
	"testing"
	"tetris"
	"tetris/combo4"
)

func BenchmarkNextState(b *testing.B) {
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	states := nfa.States().Slice()

	p := PolicyFromScorer(nfa, NewNFAScorer(nfa, 7))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		randState := states[rand.Intn(len(states))]
		queue := tetris.RandPieces(7)
		var bag tetris.PieceSet

		p.NextState(randState, queue[0], queue[1:], bag)
	}
}

func testPolicySucessRate(t *testing.T, p Policy, want float64) {
	const (
		trials         = 100
		piecesPerTrial = 100
	)
	rand.Seed(110)

	var incomplete int
	for t := 0; t < trials; t++ {
		queue := tetris.RandPieces(piecesPerTrial)
		input := make(chan tetris.Piece, 1)
		output := StartGame(p, combo4.LeftI, queue[0], queue[1:7], input)
		for _, p := range queue[7:] {
			input <- p
			if <-output == nil {
				incomplete++
				break
			}
		}
	}
	if ratio := 1 - float64(incomplete)/trials; ratio < want {
		t.Errorf("Decider has win rate=%.2f, want at least %.2f", ratio, want)
	}
}

func TestNFASucessRate(t *testing.T) {
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	testPolicySucessRate(t, PolicyFromScorer(nfa, NewNFAScorer(nfa, 7)), 0.7)
}
