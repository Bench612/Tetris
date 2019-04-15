// Package bot makes decisions on playing 4 wide combos.
package bot

import (
	"sort"
	"tetris"
	"tetris/combo4"
)

// Scorer scores a sitaution on how good it is.
type Scorer interface {
	// A higher score means the situation is better than others.
	Score(state combo4.State, next []tetris.Piece, bagUsed tetris.PieceSet) int64
}

// NFAScorer gives scores for situtations based on the number of permutations of
// that have a possible solution i.e situations that an NFA considers doable.
type NFAScorer struct {
	nfa *combo4.NFA
	// The length of permutations considered. A larger permLen leads to more
	// accurate scores.
	permLen int
	// Store the permutations of permLen that will fail for each state.
	inviable map[combo4.State]*tetris.SeqSet
	// Precompute the size of each inviable SeqSet for each state.
	inviableSizes map[combo4.State]int
}

// Score looks at the next pieces and all permutations of length permLen after
// the next pieces and sees which ones an NFA could solve.
func (s *NFAScorer) Score(state combo4.State, next []tetris.Piece, bagUsed tetris.PieceSet) int64 {
	tuple := s.scoreTuple(state, next, bagUsed)

	// Score by (in order of importance)
	// 1) The number of elements consumed. (must be less than 2^13=8192)
	// 2) The viable/inviable permutations (must be less than 2^40)
	// 3) The number of states.            (must be less than 2^10=1024)
	return int64(tuple.consumed<<50) - int64(tuple.invalidPermutations<<10) + int64(tuple.numStates)
}

type scoreTuple struct {
	consumed            int
	invalidPermutations int
	numStates           int
}

func (s *NFAScorer) scoreTuple(state combo4.State, next []tetris.Piece, bagUsed tetris.PieceSet) scoreTuple {
	endStates, consumed := s.nfa.EndStates(combo4.NewStateSet(state), next)

	score := scoreTuple{
		consumed:  consumed,
		numStates: len(endStates),
	}
	if consumed == len(next) {
		score.invalidPermutations = s.inviableSeqs(endStates, bagUsed)
	}
	return score
}

func (s *NFAScorer) inviableSeqs(endStates combo4.StateSet, bagUsed tetris.PieceSet) int {
	// Try the states with the least failures first to reduce the set.
	states := endStates.Slice()
	sort.Slice(states, func(i, j int) bool { return s.inviableSizes[states[i]] < s.inviableSizes[states[j]] })

	inviableForAll := tetris.Permutations(bagUsed)
	for _, state := range states {
		inviableForAll = inviableForAll.Intersection(s.inviable[state])
	}
	// Score by the number of inviable sequences.
	return inviableForAll.Size(s.permLen)
}

type stateInviable struct {
	state    combo4.State
	inviable *tetris.SeqSet
}

// NewNFAScorer creates a new Scorer based on permutations of the specified length.
func NewNFAScorer(nfa *combo4.NFA, permLen int) *NFAScorer {
	states := nfa.States().Slice()
	if len(states) > 2<<10 {
		panic("Too many possible states to generate a score")
	}

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
	return &NFAScorer{
		nfa:           nfa,
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
