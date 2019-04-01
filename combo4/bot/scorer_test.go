package bot

import (
	"testing"
	"tetris"
	"tetris/combo4"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			desc: "Five bag (J,S remaining)",
			bag:  tetris.NewPieceSet(tetris.J, tetris.S).Inverted(),
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

func TestAllBags(t *testing.T) {
	bags := allBags()
	seen := make(map[tetris.PieceSet]bool)
	for _, b := range bags {
		if seen[b] {
			t.Errorf("bag %v is duplicated", b)
		}
		seen[b] = true
	}
	if len(bags) != 128 { // 2^7
		t.Errorf("got %d bags, want 128", len(bags))
	}
}

func TestEncodeDecode(t *testing.T) {
	pieceSet := tetris.NewPieceSet(tetris.S)
	state := combo4.State{Field: combo4.LeftI, Hold: tetris.L}
	seq := tetris.MustPieceSeq(tetris.NonemptyPieces[:])
	s := &Scorer{
		inviable: map[tetris.PieceSet]map[combo4.State]map[tetris.PieceSeq]bool{
			pieceSet: map[combo4.State]map[tetris.PieceSeq]bool{
				state: map[tetris.PieceSeq]bool{
					seq: true,
				},
			},
		},
	}
	bytes, err := s.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode failed: %v", err)
	}

	got := &Scorer{}
	if err := got.GobDecode(bytes); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	if diff := cmp.Diff(s.inviable, got.inviable, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("mismatch after encoding + decoding (-want +got):\n%s", diff)
	}
}
