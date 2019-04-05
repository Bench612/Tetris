// Package bot creates a program to play 4 wide combos.
package bot

import (
	"bytes"
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
	// Store the size of each sequence set for each state.
	sizes map[combo4.State]int
	// Store the possible sequences of length 7 for each bag state
	seq7 [255]*tetris.SeqSet
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
	// Try the states with the least failures first to reduce the set.
	states := stateSet.Slice()
	sort.Slice(states, func(i, j int) bool { return s.sizes[states[i]] < s.sizes[states[j]] })

	inviableForAll := s.seq7[bagUsed]
	for _, state := range states {
		inviableForAll = inviableForAll.Intersection(s.inviable[state])
	}
	return 5040 - inviableForAll.Size(7)
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

// genScorer can be used to generate a new Scorer without the raw bytes.
func genScorer() *Scorer {
	s := &Scorer{}

	s.inviable = make(map[combo4.State]*tetris.SeqSet)
	nfa, states := continuousNFAAndStates()
	for _, state := range states {
		s.inviable[state] = genInviableSeqs(nfa, state)
	}

	s.sizes = make(map[combo4.State]int, len(s.inviable))
	for state, seqSet := range s.inviable {
		s.sizes[state] = seqSet.Size(7)
	}

	for _, bag := range allBags() {
		seqs := new(tetris.SeqSet)
		forEach7Seq(bag, func(seq []tetris.Piece) {
			seqs.AddPrefix(seq)
		})
		s.seq7[bag] = seqs
	}
	return s
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
	if err := decoder.Decode(&s.inviable); err != nil {
		return fmt.Errorf("gob.Decode(inviable): %v", err)
	}
	if err := decoder.Decode(&s.sizes); err != nil {
		return fmt.Errorf("gob.Decode(sizes): %v", err)
	}
	if err := decoder.Decode(&s.seq7); err != nil {
		return fmt.Errorf("gob.Decode(seq7): %v", err)
	}
	return nil
}

// GobEncode returns a bytes representation of the Scorer.
func (s *Scorer) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&s.inviable); err != nil {
		return nil, fmt.Errorf("encoder.Encode(inviable): %v", err)
	}
	if err := encoder.Encode(&s.sizes); err != nil {
		return nil, fmt.Errorf("encoder.Encode(sizes): %v", err)
	}
	if err := encoder.Encode(&s.seq7); err != nil {
		return nil, fmt.Errorf("encoder.Encode(seq7): %v", err)
	}
	return buf.Bytes(), nil
}
