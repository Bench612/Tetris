// Package bot creates a program to play 4 wide combos.
package bot

import (
	"tetris"
	"tetris/combo4"
)

// Scorer provides a score for how good a situation is.
type Scorer interface {
	Score(states combo4.StateSet, bag tetris.PieceSet) int
}

// Seq7Scorer gives scores for situtations based on the number of 7 pieces
// permutations that have a possible solutiuon.
type Seq7Scorer struct {
	// Store the sequences that will fail for each state.
	inviable map[combo4.State]map[tetris.PieceSeq]bool
}

// NewSeq7Scorer creates a new Seq7Scorer.
func NewSeq7Scorer() *Seq7Scorer {
	moves := combo4.AllContinuousMoves()
	nfa := combo4.NewNFA(moves)

	fields := make(map[combo4.Field4x4]bool)
	for _, m := range moves {
		fields[m.Start] = true
	}
	allPieces := append([]tetris.Piece{tetris.EmptyPiece}, tetris.NonemptyPieces[:]...)
	states := make([]combo4.State, 0, len(fields)*len(allPieces))
	for f := range fields {
		for _, h := range allPieces {
			states = append(states, combo4.State{
				Field: f,
				Hold:  h,
			})
		}
	}

	bag := tetris.NewPieceSet()
	inviable := make(map[combo4.State]map[tetris.PieceSeq]bool, len(states))
	for _, state := range states {
		inviableSeqs := make(map[tetris.PieceSeq]bool)
		forEach7Seq(bag, func(perm []tetris.Piece) {
			if len(nfa.EndStates(state, perm...)) == 0 {
				inviableSeqs[tetris.MustPieceSeq(perm)] = true
			}
		})
		inviable[state] = inviableSeqs
	}

	return &Seq7Scorer{
		inviable: inviable,
	}
}

// Score provides a score for how good a situation.
func (*Seq7Scorer) Score(states combo4.StateSet, bagUsed tetris.PieceSet) int {
	// Not implemented yet.
	return 0
}

// forEach7Seq calls the do() func for each possible sequence of 7 pieces.
// An empty bag states represents that all pieces are possible.
func forEach7Seq(used tetris.PieceSet, do func(permutation []tetris.Piece)) {
	unused := used.Inverted()
	unusedLen := unused.Len()
	pieces := append(unused.ToSlice(), tetris.NonemptyPieces[:]...)
	forEachPerm(pieces[:unusedLen], func() {
		forEachPermEarlyStop(pieces[unusedLen:], unusedLen, func() {
			do(pieces[:7])
		})
	})
}

// forEachPerm rearranges the slice and calls do() for every possible
// permutation of the slice. When forEachPerm is finished, all elements
// are in their original locations.
func forEachPerm(p []tetris.Piece, do func()) {
	forEachPermEarlyStop(p, 1, do)
}

func forEachPermEarlyStop(p []tetris.Piece, dontPermuteLen int, do func()) {
	if len(p) <= dontPermuteLen {
		do()
		return
	}
	for swap := 0; swap < len(p); swap++ {
		p[0], p[swap] = p[swap], p[0]
		forEachPermEarlyStop(p[1:], dontPermuteLen, do)
		p[0], p[swap] = p[swap], p[0]
	}
}
