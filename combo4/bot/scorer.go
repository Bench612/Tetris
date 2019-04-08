// Package bot makes decisions on playing 4 wide combos.
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
// that have a possible solution.
type Scorer struct {
	// The length of permutations considered. A larger permLen leads to more
	// accurate scores.
	permLen int
	// Store the permutations of permLen that will fail for each state.
	inviable map[combo4.State]*tetris.SeqSet
	// Precompute the size of each inviable SeqSet for each state.
	inviableSizes map[combo4.State]int
}

// NewScorer creates a new Scorer based on permutations of length 7.
func NewScorer() *Scorer {
	s := &Scorer{}
	if err := s.GobDecode(scorerGob); err != nil {
		panic(fmt.Errorf("GobDecode failed: %v", err))
	}
	return s
}

// Score provides a score for how good a situation is.
// Higher scores are better.
func (s *Scorer) Score(stateSet combo4.StateSet, bagUsed tetris.PieceSet) int {
	// Try the states with the least failures first to reduce the set.
	states := make([]combo4.State, 0, len(stateSet))
	for s := range stateSet {
		states = append(states, s)
	}
	sort.Slice(states, func(i, j int) bool { return s.inviableSizes[states[i]] < s.inviableSizes[states[j]] })

	inviableForAll := tetris.Permutations(bagUsed)
	for _, state := range states {
		inviableForAll = inviableForAll.Intersection(s.inviable[state])
	}
	// Score by the number of inviable sequences. Tie break by the number of states.
	scoreMajor := (-inviableForAll.Size(s.permLen))
	scoreMinor := len(stateSet)
	// Shift the major score by 10 bits. This assumes the minor score is always less than 2^10.
	return scoreMajor<<10 + scoreMinor
}

// GenScorer can be used to generate a new Scorer without the raw byte encoding.
// Generally NewScorer() should be used.
func GenScorer(permutationLen int) *Scorer {
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
		permLen:       permutationLen,
		inviable:      inviable,
		inviableSizes: genSizes(inviable, permutationLen),
	}
}

func genSizes(inviable map[combo4.State]*tetris.SeqSet, permLen int) map[combo4.State]int {
	sizes := make(map[combo4.State]int, len(inviable))
	for state, seqSet := range inviable {
		sizes[state] = seqSet.Size(permLen)
	}
	return sizes
}

// genInviableSeqs generates the inviable sequences of length 7 for the given
// state.
func genInviableSeqs(nfa *combo4.NFA, state combo4.State) *tetris.SeqSet {
	inviable := new(tetris.SeqSet)
	for _, bag := range tetris.AllPieceSets() {
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
	if err := decoder.Decode(&s.permLen); err != nil {
		return fmt.Errorf("gob.Decode(permLen): %v", err)
	}
	if err := decoder.Decode(&s.inviable); err != nil {
		return fmt.Errorf("gob.Decode(inviable): %v", err)
	}
	s.inviableSizes = genSizes(s.inviable, s.permLen)
	return nil
}

// GobEncode returns a bytes representation of the Scorer.
func (s *Scorer) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&s.permLen); err != nil {
		return nil, fmt.Errorf("encoder.Encode(permLen): %v", err)
	}
	if err := encoder.Encode(&s.inviable); err != nil {
		return nil, fmt.Errorf("encoder.Encode(inviable): %v", err)
	}
	return buf.Bytes(), nil
}
