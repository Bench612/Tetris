package bot

import (
	"bytes"
	"math/rand"
	"testing"
	"tetris"
	"tetris/combo4"
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
				seen := make(map[tetris.Seq]bool)
				forEach7Seq(test.bag, func(perm []tetris.Piece) {
					if len(perm) != 7 {
						t.Fatalf("%v: expected 7 elements in permuation", perm)
					}
					seq := tetris.MustSeq(perm)
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
	seqSet := new(tetris.SeqSet)
	seqSet.AddPrefix(tetris.NonemptyPieces[:])
	s := &Scorer{
		inviable: map[combo4.State]*tetris.SeqSet{
			combo4.State{Field: combo4.LeftI, Hold: tetris.L}: seqSet,
		},
	}

	enc1, err := s.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode failed: %v", err)
	}

	got := &Scorer{}
	if err := got.GobDecode(enc1); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	enc2, err := s.GobEncode()
	if err != nil {
		t.Fatalf("Second GobEncode failed: %v", err)
	}

	if !bytes.Equal(enc1, enc2) {
		t.Fatalf("First and second GobEncode returned different values ")
	}
}

var scorerBenchmark *Scorer

func BenchmarkNewScorer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		scorerBenchmark = NewScorer()
	}
}

func BenchmarkScore(b *testing.B) {
	s := NewScorer()
	_, states := continuousNFAAndStates()
	set := combo4.NewStateSet()
	for len(set) < 50 {
		randIdx := rand.Intn(len(states))
		set[states[randIdx]] = true
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Score(set, tetris.NewPieceSet(tetris.I, tetris.J))
	}
}

func TestScorer(t *testing.T) {
}
