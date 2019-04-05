// Package bot creates a program to play 4 wide combos.
package bot

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"fmt"
	"sort"
	"tetris"
	"tetris/combo4"
)

// Scorer gives scores for situtations based on the number of 7 pieces
// permutations that have a possible solutiuon.
type Scorer struct {
	// Store the sequences that will fail for each state.
	inviable map[combo4.State]*tetris.SeqSet
	// Store the size of each sequence set for each bag state.
	sizes map[combo4.State]int
}

// NewScorer creates a new Scorer.
func NewScorer() *Scorer {
	s := &Scorer{}
	if err := s.GobDecode(scorerGob); err != nil {
		panic(fmt.Errorf("GobDecode failed: %v", err))
	}
	return s
}

// Score provides a score for how good a situation is.
func (s *Scorer) Score(stateSet combo4.StateSet, bagUsed tetris.PieceSet) int {
	if len(stateSet) == 0 {
		return 0
	}

	// Try the states with the least failures first.
	states := stateSet.Slice()
	sort.Slice(states, func(i, j int) bool { return s.sizes[states[i]] < s.sizes[states[j]] })

	// Keep track of sequences that are inviable for every state.
	inviableSeqs := list.New()
	forEach7Seq(bagUsed, func(seq []tetris.Piece) {
		if !s.inviable[states[0]].Contains(seq) {
			return
		}
		cpy := make([]tetris.Piece, 7)
		copy(cpy, seq)
		inviableSeqs.PushBack(cpy)
	})
	for _, state := range states[1:] {
		inviable := s.inviable[state]
		for e := inviableSeqs.Front(); e != nil; {
			next := e.Next()
			if !inviable.Contains(e.Value.([]tetris.Piece)) {
				inviableSeqs.Remove(e)
			}
			e = next
		}
	}
	return 5040 - inviableSeqs.Len()
}

func continuousNFAAndStates() (*combo4.NFA, []combo4.State) {
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
	return nfa, states
}

// genInviable generates the inviable map from scratch.
func genInviable() map[combo4.State]*tetris.SeqSet {
	nfa, states := continuousNFAAndStates()

	inviable := make(map[combo4.State]*tetris.SeqSet)
	for _, state := range states {
		inviable[state] = genInviableSeqs(nfa, state)
	}
	return inviable
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

// genInviableSeqs generates the inviable sequences of length 7 for the given
// state.
func genInviableSeqs(nfa *combo4.NFA, state combo4.State) *tetris.SeqSet {
	inviable := new(tetris.SeqSet)
	for _, bag := range allBags() {
		forEach7Seq(bag, func(perm []tetris.Piece) {
			if inviable.Contains(perm) {
				return
			}
			_, unconsumed := nfa.TryConsume(combo4.NewStateSet(state), perm...)
			if len(unconsumed) != 0 {
				prefixLen := len(perm) - len(unconsumed) + 1
				inviable.AddPrefix(perm[:prefixLen])
			}
		})
	}
	return inviable
}

// forEach7Seq calls the do() func for each possible sequence of 7 pieces given
// the state of the bag. An empty bag state represents that no pieces have been
// played.
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
	var inviable map[combo4.State]*tetris.SeqSet
	if err := decoder.Decode(&inviable); err != nil {
		return fmt.Errorf("gob.Decode: %v", err)
	}
	s.inviable = inviable

	s.sizes = make(map[combo4.State]int, len(inviable))
	for state, seqSet := range inviable {
		s.sizes[state] = seqSet.Size(7)
	}

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
