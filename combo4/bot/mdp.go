package bot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"sync"
	"tetris"
	"tetris/combo4"
	"time"
)

// MDP represents a Markov Decision Process but only considers the game states
// that are considered "stable". That is, states with a piece held and are not
// swap restricted.
type MDP struct {
	nfa        *combo4.NFA
	previewLen int

	// A map from GameState to the next chosen state.
	policy map[GameState]combo4.State

	// The expected value for how many combos will occur minus previewLen.
	// With error margin up to 1.
	value map[GameState]int
}

// GameState encapsulates all information about the current game state while
// doing 4 wide combos. GameState can be used as map key.
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

	m := &MDP{
		nfa:        combo4.NewNFA(combo4.AllContinuousMoves()),
		previewLen: previewLen,
		value:      make(map[GameState]int, int(128*28*7*7*math.Pow(2.6, float64(previewLen)))),
	}

	var filteredStates []combo4.State
	for state := range m.nfa.States() {
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
						if m.isStable(gState) {
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
		m.value[gState] = 1
	}

	m.initPolicy()
	return m, nil
}

// initPolicy creates an initial policy based on what the scorer would do.
// initPolicy assumes the scores have been initialized.
func (m *MDP) initPolicy() {
	m.policy = make(map[GameState]combo4.State, len(m.value))
	d := PolicyFromScorer(m.nfa, NewNFAScorer(m.nfa, m.previewLen))
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

// UpdatePolicy updates the policy based on values and returns whether
// it has changed.
func (m *MDP) UpdatePolicy() bool {
	var changed int

	for gState, currentChoice := range m.policy {
		choices := m.nfa.NextStates(gState.State, gState.Current)
		if len(choices) == 1 {
			m.policy[gState] = choices[0]
		}

		// Use the current choice in case of a tie-break.
		// This way the policy changes minimally.
		bestVal := m.calcValue(gState, currentChoice)
		bestChoice := currentChoice
		for _, choice := range choices {
			if choice == currentChoice {
				continue
			}
			if v := m.calcValue(gState, choice); v > bestVal {
				bestVal = v
				bestChoice = choice
			}
		}

		if currentChoice != bestChoice {
			changed++
			m.policy[gState] = bestChoice
			// Update the corresponding values to the new estimates.
			m.value[gState] = bestVal
		}
	}
	log.Printf("Updated policy with %d changes", changed)
	return changed != 0
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
	return policy, NewNFAScorer(m.nfa, m.previewLen)
}

type valueChange struct {
	gState     GameState
	policy     combo4.State
	valCurrent int
	valChange  int // Should always be positive
}

// UpdateValues updates the expected values based on the current
// expected values and policy. UpdateValues returns whether anything got
// changed.
func (m *MDP) UpdateValues() bool {
	toUpdate := make([]valueChange, 0, len(m.value))
	for gState, v := range m.value {
		toUpdate = append(toUpdate, valueChange{
			gState:     gState,
			policy:     m.policy[gState],
			valCurrent: v,
			valChange:  0,
		})
	}

	var hasChanged bool
	for iter := 0; len(toUpdate) > 0; iter++ {

		sort.Slice(toUpdate, func(i, j int) bool {
			// Prioritize high value changes.
			if toUpdate[i].valChange != toUpdate[j].valChange {
				return toUpdate[i].valChange > toUpdate[j].valChange
			}
			// Prioritize small current values.
			return toUpdate[i].valCurrent < toUpdate[j].valCurrent
		})
		if toUpdate[len(toUpdate)-1].valChange < 0 {
			panic("value decreased. should never happen")
		}

		var changes int
		for idx := range toUpdate {
			c := &toUpdate[idx]

			new := m.calcValue(c.gState, c.policy)
			if old := c.valCurrent; new != old {
				changes++
				m.value[c.gState] = new
				c.valCurrent = new
				c.valChange = new - old // Difference should always be positive
			}
		}
		log.Printf("Updated %d values (#%d)", changes, iter)
		if changes == 0 {
			break
		}
		hasChanged = true
	}
	return hasChanged
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
	return 1 + (totalSubExp)/len(possibleNextPiece)
}

// Update updates the MDP until it is at an optimal policy while periodically
// saving progress to the given filePath.
func (m *MDP) Update(filePath string) error {
	for i := 0; ; i++ {
		start := time.Now()
		valueChanges := m.UpdateValues()
		log.Printf("UpdatedValues (#%d) in %v", i, time.Since(start))
		if valueChanges {
			if err := m.save(filePath); err != nil {
				return fmt.Errorf("save failed: %v", err)
			}
		} else {
			return nil
		}

		start = time.Now()
		policyChanges := m.UpdatePolicy()
		log.Printf("UpdatePolicy (#%d) in %v", i, time.Since(start))
		if !policyChanges {
			return nil
		}
	}
}

// save saves the MDP to the filePath or returns nil
// if the path is empty.
func (m *MDP) save(filePath string) error {
	if filePath == "" {
		return nil
	}

	start := time.Now()
	bytes, err := m.GobEncode()
	if err != nil {
		return fmt.Errorf("encode failed: %v", err)
	}
	if err := ioutil.WriteFile(filePath, []byte(bytes), 0644); err != nil {
		return fmt.Errorf("WriteFile failed: %v", err)
	}
	log.Printf("Updated file in %v\n", time.Since(start))
	return nil
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
	m.initPolicy()
	m.UpdatePolicy()
	return nil
}
