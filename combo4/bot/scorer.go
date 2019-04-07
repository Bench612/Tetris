// Package bot makes decisions on playing to play 4 wide combos.
package bot

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"tetris"
	"tetris/combo4"
)

// Scorer gives scores for situtations based on the number of permutations of
// length 7 that have a possible solution.
type Scorer struct {
	// Store the permutations that will fail for each state.
	inviable map[combo4.State]*tetris.SeqSet
	// Store the size of each SeqSet for each state.
	sizes map[combo4.State]int
	// Store the possible permutations for each bag state
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
func (s *Scorer) Score(stateSet combo4.StateSet, bagUsed tetris.PieceSet) int32 {
	// Try the states with the least failures first to reduce the set.
	states := make([]combo4.State, 0, len(stateSet))
	for s := range stateSet {
		states = append(states, s)
	}
	sort.Slice(states, func(i, j int) bool { return s.sizes[states[i]] < s.sizes[states[j]] })

	inviableForAll := s.seq7[bagUsed]
	for _, state := range states {
		inviableForAll = inviableForAll.Intersection(s.inviable[state])
	}
	// Each prefix will be length 7 so the size is also the number of sequences.
	validSeqs := 5040 - int32(inviableForAll.Size(7))
	// Tie break if the number of validSeqs are equal by the size of the state set.
	return validSeqs<<10 + int32(len(stateSet))
}

// GenScorer can be used to generate a new Scorer without the raw byte encoding.
// Generally NewScorer() should be used.
func GenScorer() *Scorer {
	inviable := make(map[combo4.State]*tetris.SeqSet)
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())

	type queueItem struct {
		state    combo4.State
		inviable *tetris.SeqSet
	}
	queue := make(chan queueItem, 16)
	concurrency := make(chan bool, 16) // Limit concurrency to 16 routines.
	states := nfa.States()
	for state := range states {
		state := state // Capture range variable.
		go func() {
			concurrency <- true
			queue <- queueItem{state, genInviableSeqs(nfa, state)}
			<-concurrency
		}()
	}
	for range states {
		item := <-queue
		inviable[item.state] = item.inviable
	}

	return &Scorer{
		inviable: inviable,
		sizes:    genSizes(inviable),
		seq7:     genSeq7(),
	}
}

func genSizes(inviable map[combo4.State]*tetris.SeqSet) map[combo4.State]int {
	sizes := make(map[combo4.State]int, len(inviable))
	for state, seqSet := range inviable {
		sizes[state] = seqSet.Size(7)
	}
	return sizes
}

func genSeq7() [255]*tetris.SeqSet {
	var seq7 [255]*tetris.SeqSet
	for _, bag := range allBags() {
		seqs := new(tetris.SeqSet)
		forEach7Seq(bag, func(seq []tetris.Piece) {
			seqs.AddPrefix(seq)
		})
		seq7[bag] = seqs
	}
	return seq7
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
			if _, consumed := nfa.EndStates(combo4.NewStateSet(state), perm); consumed != len(perm) {
				inviable.AddPrefix(perm[:consumed+1])
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
