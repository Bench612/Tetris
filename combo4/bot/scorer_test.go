package bot

import (
	"testing"
	"tetris"
	"tetris/combo4"
)

func BenchmarkNewNFAScorer7(b *testing.B) {
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	for n := 0; n < b.N; n++ {
		_ = NewNFAScorer(nfa, 7)
	}
}

func BenchmarkNewNFAScorer8(b *testing.B) {
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	for n := 0; n < b.N; n++ {
		_ = NewNFAScorer(nfa, 8)
	}
}

func TestInviableSeqs(t *testing.T) {
	tests := []struct {
		desc   string
		states combo4.StateSet
		bag    tetris.PieceSet
	}{
		{
			desc:   "One state, empty bag",
			states: combo4.NewStateSet(combo4.State{Field: combo4.LeftI}),
		},
		{
			desc: "Two states, I,J bag",
			states: combo4.NewStateSet(
				combo4.State{Field: combo4.LeftI, Hold: tetris.J},
				combo4.State{Field: combo4.RightI, Hold: tetris.I}),
			bag: tetris.NewPieceSet(tetris.I, tetris.J),
		},
	}
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	s := NewNFAScorer(nfa, 7)
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var want int
			forEachSeq(test.bag, 7, func(seq []tetris.Piece) {
				if _, consumed := nfa.EndStates(test.states, seq); consumed != s.permLen {
					want++
				}
			})

			if got := s.inviableSeqs(test.states, test.bag); got != want {
				t.Errorf("got inviableSeqs()=%d, want %d", got, want)
			}
		})
	}
}
