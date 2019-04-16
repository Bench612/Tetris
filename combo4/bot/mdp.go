package bot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"tetris"
	"tetris/combo4"
)

// MDP represents a Markov Decision Process but only considers the game states
// that are considered "stable". That is, states with a piece held and are not
// swap restricted.
type MDP struct {
	nfa    *combo4.NFA
	scorer *NFAScorer

	previewLen int

	// A map from GameState to the next chosen state.
	policy map[GameState]combo4.State

	// The expected value for how many combos will occur minus previewLen.
	// With error margin up to 1.
	value map[GameState]int

	// Used for optimization internally.
	recentlyChanged []valueChange
}

// GameState encapsulates all information about the current game state while
// doing 4 wide combos.
type GameState struct {
	State   combo4.State
	Current tetris.Piece
	Preview tetris.Seq
	BagUsed tetris.PieceSet
}

type valueChange struct {
	gState GameState
	change int
}

// NewMDP constructs a new MDP.
func NewMDP(previewLen int) (*MDP, error) {
	if previewLen > 7 || previewLen < 1 {
		return nil, errors.New("previewLen must be between 1 and 7")
	}
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())

	prealloc := int(128 * 28 * 7 * 7 * math.Pow(2.6, float64(previewLen)))
	mdp := &MDP{
		nfa:             nfa,
		scorer:          NewNFAScorer(nfa, previewLen),
		previewLen:      previewLen,
		value:           make(map[GameState]int, prealloc),
		recentlyChanged: make([]valueChange, 0, prealloc),
	}

	var filteredStates []combo4.State
	for state := range nfa.States() {
		// Don't include states that usually only show up in the beginning.
		if state.SwapRestricted || state.Hold == tetris.EmptyPiece {
			continue
		}
		filteredStates = append(filteredStates, state)
	}

	stableCh := make(chan GameState, 5000)
	go func() {
		allBags := tetris.AllPieceSets()
		var wg sync.WaitGroup
		wg.Add(len(allBags))
		maxConcurrency := make(chan bool, 4)
		for _, bagUsed := range allBags {
			bagUsed := bagUsed // Capture range variable.

			maxConcurrency <- true
			go func() {
				defer func() { <-maxConcurrency }()
				defer wg.Done()

				reversed := make([]tetris.Piece, previewLen+1)
				forEachSeq(bagUsed.Inverted(), previewLen+1, func(seq []tetris.Piece) {
					for i, p := range seq {
						reversed[len(reversed)-1-i] = p
					}
					current := reversed[0]
					preview := tetris.MustSeq(reversed[1:])
					for _, state := range filteredStates {
						gState := GameState{
							State:   state,
							Current: current,
							Preview: preview,
							BagUsed: bagUsed,
						}
						if mdp.isStable(gState) {
							stableCh <- gState
						}
					}
				})
			}()
		}
		wg.Wait()
		close(stableCh)
	}()

	for gState := range stableCh {
		mdp.value[gState] = 1
	}

	mdp.policy = make(map[GameState]combo4.State, len(mdp.value))
	mdp.UpdatePolicy()
	return mdp, nil
}

// initPolicy creates an initial policy based on what the scorer would do.
func (m *MDP) initPolicy() {
	m.policy = make(map[GameState]combo4.State, len(m.value))
	d := NewScoreDecider(m.nfa, m.scorer)
	for gState := range m.value {
		choice := d.NextState(gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed)
		m.policy[gState] = *choice
	}
}

// isStable is used to compute the initial values.
// A GameState is considered stable if the current + preview can be consumed.
func (m *MDP) isStable(gState GameState) bool {
	start := combo4.NewStateSet(m.nfa.NextStates(gState.State, gState.Current)...)
	_, consumed := m.nfa.EndStates(start, gState.Preview.Slice())
	return consumed == m.previewLen
}

func forEachSeq(bagUsed tetris.PieceSet, seqLen int, do func([]tetris.Piece)) {
	seq := make([]tetris.Piece, seqLen)
	forEachSeqHelper(seq, bagUsed, 0, do)
}

func forEachSeqHelper(seq []tetris.Piece, bagUsed tetris.PieceSet, seqIdx int, do func([]tetris.Piece)) {
	if bagUsed.Len() == 7 {
		bagUsed = 0
	}
	for _, p := range bagUsed.Inverted().Slice() {
		seq[seqIdx] = p
		if seqIdx == len(seq)-1 {
			do(seq)
			continue
		}
		forEachSeqHelper(seq, bagUsed.Add(p), seqIdx+1, do)
	}
}

// UpdatePolicy updates the policy based on values and returns the
// number of changes.
func (m *MDP) UpdatePolicy() int {
	var changed int

	for gState := range m.value {
		choices := m.nfa.NextStates(gState.State, gState.Current)
		if len(choices) == 1 {
			m.policy[gState] = choices[0]
		}

		var bestVal int
		var bestChoice combo4.State
		for _, choice := range choices {
			if v := m.calcValue(gState, choice); v > bestVal {
				bestVal = v
				bestChoice = choice
			}
		}

		if m.policy[gState] != bestChoice {
			changed++
			m.policy[gState] = bestChoice
			m.value[gState] = bestVal
		}
	}
	return changed
}

// Policy returns the MDP's policy. The given map is used and the Scorer
// should be used if no entry in the map exists.
func (m *MDP) Policy() (map[GameState]combo4.State, Scorer) {
	policy := make(map[GameState]combo4.State, len(m.policy))
	for gState, choice := range m.policy {
		choices := m.nfa.NextStates(gState.State, gState.Current)

		// Only specify the choice if its not obvious.
		if len(choices) > 1 {
			policy[gState] = choice
		}
	}
	return policy, m.scorer
}

// UpdateValues updates the expected values based on the current
// expected values and policy. Updatevalueues returns the number of
// values with significant change.
//
// Cutoff is used to shorten time it takes to converge. Expected values
// are not calculated if they are higher than the cutoff. Use cutoff=-1 to
// find the exact values. The idea is to first stabilize the policy for all
// the lower values which are dependencies and then increase the cutoff.
func (m *MDP) UpdateValues(cutoff int) int {
	// Update the recently changed first.
	for _, c := range m.recentlyChanged {
		m.value[c.gState] = m.calcValue(c.gState, m.policy[c.gState])
	}

	if m.recentlyChanged == nil {
		m.recentlyChanged = make([]valueChange, 0, len(m.value))
	}
	m.recentlyChanged = m.recentlyChanged[:0]

	// Update all states.
	for gState, choice := range m.policy {
		before := m.value[gState]
		if before > cutoff && cutoff != -1 {
			continue
		}
		after := m.calcValue(gState, choice)
		if before != after {
			m.value[gState] = after
			abs := after - before
			if after < before {
				abs = before - after
			}
			m.recentlyChanged = append(m.recentlyChanged, valueChange{gState, abs})
		}
	}
	// Prioritize the largest changes first in the next iteration.
	sort.Slice(m.recentlyChanged, func(i, j int) bool { return m.recentlyChanged[i].change > m.recentlyChanged[j].change })
	return len(m.recentlyChanged)
}

func (m *MDP) calcValue(cur GameState, choice combo4.State) int {
	var (
		current        = cur.Preview.AtIndex(0)
		previewShifted = cur.Preview.RemoveFirst()
	)

	bag := cur.BagUsed
	if bag.Len() == 7 {
		bag = 0
	}
	possibleNextPiece := bag.Inverted().Slice()

	var totalSubExp int
	for _, p := range possibleNextPiece {
		var newBag tetris.PieceSet
		if cur.BagUsed.Len() == 7 {
			newBag = p.PieceSet()
		} else {
			newBag = bag.Add(p)
		}

		next := GameState{
			State:   choice,
			Current: current,
			Preview: previewShifted.SetIndex(m.previewLen-1, p),
			BagUsed: newBag,
		}
		totalSubExp += m.value[next]
	}
	return 1 + totalSubExp/len(possibleNextPiece)
}

// GobEncode returns a Gob encoding of a MDP.
func (m *MDP) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&m.previewLen); err != nil {
		return nil, fmt.Errorf("encoder.Encode(previewLen): %v", err)
	}
	if err := encoder.Encode(&m.value); err != nil {
		return nil, fmt.Errorf("encoder.Encode(value): %v", err)
	}
	return buf.Bytes(), nil
}

// GobDecode decodes a Gob encoding into an MDP.
func (m *MDP) GobDecode(b []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(b) // Always returns nil.
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&m.previewLen); err != nil {
		return fmt.Errorf("gob.Decode(previewLen): %v", err)
	}
	if err := decoder.Decode(&m.value); err != nil {
		return fmt.Errorf("gob.Decode(value): %v", err)
	}
	m.nfa = combo4.NewNFA(combo4.AllContinuousMoves())
	m.scorer = NewNFAScorer(m.nfa, m.previewLen)
	m.policy = make(map[GameState]combo4.State, len(m.value))
	m.UpdatePolicy()
	return nil
}
