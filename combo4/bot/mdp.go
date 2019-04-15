package bot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"sync"
	"tetris"
	"tetris/combo4"
)

// MDP represents a Markov Decision Process but only considers the game states
// that are considered "stable".
type MDP struct {
	nfa    *combo4.NFA
	scorer *NFAScorer

	previewLen int
	// The expected value for how many states it will continue in a "stable"
	// state including the gameState itself.
	expectVal map[GameState]float32
	// A map from GameState to the next chosen state.
	policy map[GameState]combo4.State
}

// GameState encapsulates all information about the current game state while
// doing 4 wide combos.
type GameState struct {
	State   combo4.State
	Current tetris.Piece
	Preview tetris.Seq
	BagUsed tetris.PieceSet
}

// NewMDP constructs a new MDP.
func NewMDP(previewLen int) (*MDP, error) {
	if previewLen > 7 || previewLen < 1 {
		return nil, errors.New("previewLen must be between 1 and 7")
	}
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())

	prealloc := 1000000
	if previewLen >= 5 {
		prealloc = 50000000
	}
	mdp := &MDP{
		nfa:        nfa,
		scorer:     NewNFAScorer(nfa, previewLen),
		previewLen: previewLen,
		expectVal:  make(map[GameState]float32, prealloc),
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
		mdp.expectVal[gState] = 1
	}

	mdp.policy = make(map[GameState]combo4.State, len(mdp.expectVal))
	mdp.UpdatePolicy()
	return mdp, nil
}

// isStable is used to compute the initial expectVals.
// A GameState is considered stable if all possible sequences of previewLen
// can be consumed.
func (m *MDP) isStable(gState GameState) bool {
	choices := m.nfa.NextStates(gState.State, gState.Current)
	for _, choice := range choices {
		endStates, consumed := m.nfa.EndStates(combo4.NewStateSet(choice), gState.Preview.Slice())
		if consumed != m.previewLen {
			continue
		}
		if m.scorer.inviableSeqs(endStates, gState.BagUsed) == 0 {
			return true
		}
	}
	return false
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

const epsilon = 0.01

// UpdatePolicy updates the policy based on  expected values and returns the
// number of changes
func (m *MDP) UpdatePolicy() int {
	var changed int

	for gState := range m.expectVal {
		choices := m.nfa.NextStates(gState.State, gState.Current)
		if len(choices) == 1 {
			m.policy[gState] = choices[0]
		}

		var bestExp float32
		var bestChoice combo4.State
		for _, choice := range choices {
			if exp := m.calcExpectValue(gState, choice); exp+epsilon > bestExp {
				bestExp = exp
				bestChoice = choice
			}
		}

		if m.policy[gState] != bestChoice {
			changed++
			m.policy[gState] = bestChoice
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

// UpdateExpectedValues updates the expected values based on the current
// expected values and policy. UpdateExpectedValues returns the number of
// values with significant change.
func (m *MDP) UpdateExpectedValues() int {
	var changed int
	for gState, choice := range m.policy {
		before := m.expectVal[gState]
		after := m.calcExpectValue(gState, choice)
		if math.Abs(float64(before-after)) > epsilon {
			changed++
			m.expectVal[gState] = after
		}
	}
	return changed
}

func (m *MDP) calcExpectValue(cur GameState, choice combo4.State) float32 {
	var (
		current        = cur.Preview.AtIndex(0)
		previewShifted = cur.Preview.RemoveFirst()
	)

	bag := cur.BagUsed
	if bag.Len() == 7 {
		bag = 0
	}
	possibleNextPiece := bag.Inverted().Slice()

	var totalSubExp float32
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
		totalSubExp += m.expectVal[next]
	}
	return 1 + totalSubExp/float32(len(possibleNextPiece))
}

// GobEncode returns a Gob encoding of a MDP.
func (m *MDP) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&m.previewLen); err != nil {
		return nil, fmt.Errorf("encoder.Encode(previewLen): %v", err)
	}
	if err := encoder.Encode(&m.expectVal); err != nil {
		return nil, fmt.Errorf("encoder.Encode(expectVal): %v", err)
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
	if err := decoder.Decode(&m.expectVal); err != nil {
		return fmt.Errorf("gob.Decode(expectVal): %v", err)
	}
	m.nfa = combo4.NewNFA(combo4.AllContinuousMoves())
	m.scorer = NewNFAScorer(m.nfa, m.previewLen)
	m.policy = make(map[GameState]combo4.State, len(m.expectVal))
	m.UpdatePolicy()
	return nil
}
