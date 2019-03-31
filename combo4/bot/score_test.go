package bot

import (
	"testing"
	"tetris"
)

func TestForEach7Seq(t *testing.T) {
	tests := []struct {
		desc      string
		bag       tetris.PieceSet
		wantCount int
	}{
		{
			desc: "Four bag (J,S,I remaining)",
			bag:  tetris.NewPieceSet(tetris.J, tetris.S, tetris.I).Inverted(),
		},
		{
			desc: "Empty bag",
		},
		{
			desc: "Full bag",
			bag:  tetris.NewPieceSet(tetris.NonemptyPieces[:]...),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			// Use t.Fatal checks to prevent spamming the error logs.
			t.Run("Basic checks", func(t *testing.T) {
				seen := make(map[tetris.PieceSeq]bool)
				forEach7Seq(test.bag, func(perm []tetris.Piece) {
					if len(perm) != 7 {
						t.Fatalf("%v: expected 7 elements in permuation", perm)
					}
					seq := tetris.MustPieceSeq(perm)
					if seen[seq] {
						t.Fatalf("%v: permutation is repeated", perm)
					}
					seen[seq] = true
				})
				if len(seen) != 5040 {
					t.Errorf("got %d unique permutations, want 5040", len(seen))
				}
			})
			// Use t.Fatal checks to prevent spamming the error logs.
			t.Run("Contains right pieces", func(t *testing.T) {
				invertLen := test.bag.Inverted().Len()
				forEach7Seq(test.bag, func(perm []tetris.Piece) {
					firstSet := tetris.NewPieceSet(perm[:invertLen]...)
					if firstSet != test.bag.Inverted() {
						t.Fatalf("%v: got %v in first %d elements, want %v", perm, firstSet, test.bag.Inverted().Len(), test.bag.Inverted())
					}

					secondSet := tetris.NewPieceSet(perm[invertLen:]...)
					if secondSet.Len() != test.bag.Len() {
						t.Fatalf("%v: expected all unique pieces in the last %d elements, got %d unique elements", perm, test.bag.Len(), secondSet.Len())
					}
				})
			})
		})
	}
}
