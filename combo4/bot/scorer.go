// Package bot creates a program to play 4 wide combos.
package bot

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"tetris"
	"tetris/combo4"
)

// Scorer gives scores for situtations based on the number of 7 pieces
// permutations that have a possible solutiuon.
type Scorer struct {
	// Store the sequences that will fail for each bag and state.
	inviable map[tetris.PieceSet]map[combo4.State]map[tetris.PieceSeq]bool
}

// NewScorer creates a new Scorer from scratch.
func NewScorer() *Scorer {
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

	inviable := make(map[tetris.PieceSet]map[combo4.State]map[tetris.PieceSeq]bool)
	for _, bag := range allBags() {
		inviableStates := make(map[combo4.State]map[tetris.PieceSeq]bool, len(states))
		for _, state := range states {
			inviableStates[state] = inviableSeqs(nfa, state, bag)
		}
		inviable[bag] = inviableStates
	}

	return &Scorer{inviable: inviable}
}

// allBags returns a list of all possible bag states.
func allBags() []tetris.PieceSet {
	bags := make([]tetris.PieceSet, 128) // 2^7
	for idx := range bags {
		var ps tetris.PieceSet
		for pieceIdx, piece := range tetris.NonemptyPieces {
			if idx&(1<<uint(pieceIdx)) != 0 {
				ps = ps.Add(piece)
			}
		}
		bags[idx] = ps
	}
	return bags
}

// inviableSeqs returns the inviable sequences of length 7 for the given state
// and bag state.
func inviableSeqs(nfa *combo4.NFA, state combo4.State, bag tetris.PieceSet) map[tetris.PieceSeq]bool {
	inviable := make(map[tetris.PieceSeq]bool)
	forEach7Seq(bag, func(perm []tetris.Piece) {
		if len(nfa.EndStates(state, perm...)) == 0 {
			inviable[tetris.MustPieceSeq(perm)] = true
		}
	})
	return inviable
}

// Score provides a score for how good a situation.
func (*Scorer) Score(states combo4.StateSet, bagUsed tetris.PieceSet) int {
	// Not implemented yet.
	return 0
}

// forEach7Seq calls the do() func for each possible sequence of 7 pieces given
// the state of the bag. An empty bag state represents that no pieces have been
//  played.
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

// GobDecode decodes bytes into the Scorer.
func (s *Scorer) GobDecode(b []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(b) // Always returns nil.
	decoder := gob.NewDecoder(buf)
	var inviable map[tetris.PieceSet]map[combo4.State]map[tetris.PieceSeq]bool
	if err := decoder.Decode(&inviable); err != nil {
		return fmt.Errorf("gob.Decode: %v", err)
	}
	s.inviable = inviable
	return nil
}

// GobEncode returns a bytes representation of the Scorer.
func (s *Scorer) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&s.inviable); err != nil {
		return nil, fmt.Errorf("encoder.Encode: %v", err)
	}
	return buf.Bytes(), nil
}
