// Package bot makes decisions on playing 4 wide combos.
package bot

import (
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

// Score provides a score for how good a situation is.
// Higher scores are better.
func (s *Scorer) Score(stateSet combo4.StateSet, bagUsed tetris.PieceSet) int {
	// Try the states with the least failures first to reduce the set.
	states := stateSet.Slice()
	sort.Slice(states, func(i, j int) bool { return s.inviableSizes[states[i]] < s.inviableSizes[states[j]] })

	inviableForAll := tetris.Permutations(bagUsed)
	for _, state := range states {
		inviableForAll = inviableForAll.Intersection(s.inviable[state])
	}
	// Score by the number of inviable sequences. Tie break by the number of states.
	return -inviableForAll.Size(s.permLen)
}

type stateInviable struct {
	state    combo4.State
	inviable *tetris.SeqSet
}

// NewScorer creates a new Scorer based on permutations of the specified length.
func NewScorer(permLen int) *Scorer {
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())
	states := nfa.States().Slice()

	ch := make(chan stateInviable, len(states))

	// Base case of all sequences of length 0 that are inviable (all viable).
	prevInviable := make(map[combo4.State]*tetris.SeqSet, len(states))
	inviable := make(map[combo4.State]*tetris.SeqSet, len(states))

	for n := 1; n <= permLen; n++ {
		prevInviable, inviable = inviable, prevInviable

		// Generate the inviable sequences of length n based on the inviable
		// sequences of length n-1.
		for _, state := range states {
			state := state // Capture range variable.
			go func() {
				var prefixToSet [8]*tetris.SeqSet
				for p := 0; p < 8; p++ {
					intersxn := tetris.ContainsAllSeqSet
					for _, endState := range nfa.NextStates(state, tetris.Piece(p)) {
						intersxn = intersxn.Intersection(prevInviable[endState])
					}
					prefixToSet[p] = intersxn
				}
				ch <- stateInviable{state, tetris.PrependedSeqSets(prefixToSet)}
			}()
		}
		for range states {
			si := <-ch
			inviable[si.state] = si.inviable
		}
	}
	return &Scorer{
		permLen:       permLen,
		inviable:      inviable,
		inviableSizes: genSizes(inviable, permLen),
	}
}

func genSizes(inviable map[combo4.State]*tetris.SeqSet, permLen int) map[combo4.State]int {
	sizes := make(map[combo4.State]int, len(inviable))
	for state, seqSet := range inviable {
		sizes[state] = seqSet.Size(permLen)
	}
	return sizes
}
