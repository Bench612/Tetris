package policy

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sync"
	"tetris"
	"tetris/combo4"
	"time"
)

// Number of go-routines to parallelize across.
const concurrency = 8

// MDP represents a Markov Decision Process but only considers the game states
// that are considered "stable". That is, states with a piece held and are not
// swap restricted.
//
// MDP is *NOT* safe for concurrent use.
type MDP struct {
	nfa        *combo4.NFA
	previewLen int

	// A map from GameState to the next chosen state.
	policy map[GameState]combo4.State

	// The expected value for how many combos will occur minus previewLen.
	// Since we only store GameStates that can at least consume the current
	// piece and all the preview, any state that it can transition to that is
	// not in the map can only consume len(preview) pieces. This is
	// conveniently the 0 value.
	value map[GameState]float64
}

// GameState encapsulates all information about the current game state while
// doing 4 wide combos. GameState can be used as map key.
type GameState struct {
	State   combo4.State
	Current tetris.Piece
	Preview tetris.Seq
	BagUsed tetris.PieceSet
}

// NewMDP constructs a new MDP for the given preview length.
func NewMDP(previewLen int) (*MDP, error) {
	if previewLen > 7 || previewLen < 0 {
		return nil, errors.New("previewLen must be between 0 and 7")
	}

	continuousMoves, _ := combo4.AllContinuousMoves()
	m := &MDP{
		nfa:        combo4.NewNFA(continuousMoves),
		previewLen: previewLen,
		value:      make(map[GameState]float64, int(128*28*7*7*math.Pow(2.6, float64(previewLen)))),
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
		maxConcurrency := make(chan bool, concurrency)
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

// ExpectedValue returns the expected number of pieces that will be consumed
// for a GameState. This is only accurate if Update() has completed.
func (m *MDP) ExpectedValue(gState GameState) float64 {
	if val, ok := m.value[gState]; ok {
		return val + float64(m.previewLen)
	}
	start := combo4.NewStateSet(m.nfa.NextStates(gState.State, gState.Current)...)
	if len(start) == 0 {
		return 0
	}
	_, consumed := m.nfa.EndStates(start, gState.Preview.Slice())
	return float64(consumed) + 1
}

// initPolicy creates an initial policy. initPolicy assumes the scores have
// been initialized.
func (m *MDP) initPolicy() {
	m.policy = make(map[GameState]combo4.State, len(m.value))
	p := FromScorer(m.nfa, NewNFAScorer(m.nfa, m.previewLen))
	for gState := range m.value {
		choice := p.NextState(gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed)
		m.policy[gState] = *choice
	}
}

// isStable is used to compute the initial values.
// A GameState is considered stable if the current + preview can be consumed.
func (m *MDP) isStable(gState GameState) bool {
	start := m.nfa.NextStates(gState.State, gState.Current)
	if len(start) == 0 {
		return false
	}
	_, consumed := m.nfa.EndStates(combo4.NewStateSet(start...), gState.Preview.Slice())
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

// updatePolicy updates the policy based on values and returns how many
// policy changes there were.
func (m *MDP) updatePolicy() int {
	var changed int
	for gState, currentChoice := range m.policy {
		choices := m.nfa.NextStates(gState.State, gState.Current)
		if len(choices) == 1 {
			continue
		}

		var (
			bestChoice combo4.State
			bestVal    = math.Inf(-1)
		)
		for _, choice := range choices {
			if v := m.calcValue(gState, choice); v > bestVal {
				bestVal = v
				bestChoice = choice
			}
		}
		if currentChoice != bestChoice {
			changed++
			m.policy[gState] = bestChoice
		}
	}
	log.Printf("Updated policy with %d changes", changed)
	return changed
}

// valueChange is used for efficient updating of values given a particular
// policy.
type valueChange struct {
	// Used to calculate the next value.
	// The next value is 1 + sum(dependencies) / possibilities
	possibilities float64
	dependencies  []*float64

	// It is important that this is updated atomically. So that valCurrent
	// never decreases. valCurrent may be concurrently modified and read.
	// But will still reach equilibrium as long as it does not decrease or
	// overly increase.
	value float64
}

const epsilon = 0.0001 // The smallest value that we care about updating.

// updateValues updates the expected values based on the current
// expected values and policy. updateValues returns the number of values
// that changed. cap can be used to specify a maximum value.
func (m *MDP) updateValues() int {
	var (
		vals    = make([]*valueChange, 0, len(m.value))
		gStates = make([]GameState, 0, len(m.value))             // Used for valueChange -> GameState
		cMap    = make(map[GameState]*valueChange, len(m.value)) // Used for GameState -> valueChange
	)
	for gState, v := range m.value {
		c := &valueChange{value: v}
		cMap[gState] = c
		vals = append(vals, c)
		gStates = append(gStates, gState)
	}
	for gState, c := range cMap {
		possibilities := m.possibilities(gState, m.policy[gState])
		for _, poss := range possibilities {
			if dep, ok := cMap[poss]; ok {
				c.dependencies = append(c.dependencies, &dep.value)
			}
		}
		c.possibilities = float64(len(possibilities))
	}
	cMap = nil // No longer needed.

	for iter := 0; ; iter++ {
		changesCh := make(chan int, 1)
		for i := 0; i < concurrency; i++ {
			start := i * len(vals) / concurrency
			end := (i + 1) * len(vals) / concurrency
			go func() {
				var changes int
				for _, c := range vals[start:end] {
					// Update val based on depdendencies.
					// Even though dependencies may change from different
					// go-routines, this is fine because it is okay to read
					// either version of the value.
					var totalVal float64
					for _, d := range c.dependencies {
						totalVal += *d
					}
					newVal := 1 + totalVal/c.possibilities

					if math.Abs(newVal-c.value) >= epsilon {
						changes++
						c.value = newVal
					}
				}
				changesCh <- changes
			}()
		}
		var changes int
		for i := 0; i < concurrency; i++ {
			changes += <-changesCh
		}
		log.Printf("Updated %d values (#%d)", changes, iter)
		if changes == 0 {
			break
		}
	}

	// Update the values map.
	var totalChanges int
	for idx, c := range vals {
		gState := gStates[idx]

		old := m.value[gState]
		if old != c.value {
			m.value[gState] = c.value
			totalChanges++
		}
	}

	return totalChanges
}

func (m *MDP) possibilities(cur GameState, choice combo4.State) []GameState {
	var (
		current        = cur.Preview.AtIndex(0)
		previewShifted = cur.Preview.RemoveFirst()
	)

	bag := cur.BagUsed
	if bag.Len() == 7 {
		bag = 0
	}
	possibleNextPiece := bag.Inverted().Slice()
	possibilities := make([]GameState, 0, len(possibleNextPiece))
	for _, p := range possibleNextPiece {
		var newBag tetris.PieceSet
		if cur.BagUsed.Len() == 7 {
			newBag = p.PieceSet()
		} else {
			newBag = bag.Add(p)
		}

		var preview tetris.Seq
		if m.previewLen > 0 {
			preview = previewShifted.SetIndex(m.previewLen-1, p)
		}

		possibilities = append(possibilities, GameState{
			State:   choice,
			Current: current,
			Preview: preview,
			BagUsed: newBag,
		})
	}
	return possibilities
}

// calcValue calculates the expected value given the current estimates and
// policy. This needs to be kept in sync with the formula in updateValues().
func (m *MDP) calcValue(cur GameState, choice combo4.State) float64 {
	var totalVal float64
	poss := m.possibilities(cur, choice)
	for _, next := range poss {
		totalVal += m.value[next]
	}
	return 1 + totalVal/float64(len(poss))
}

// Update updates the MDP until it is at an optimal policy while periodically
// saving progress to the given filePath.
func (m *MDP) Update(filePath string) error {
	for i := 0; ; i++ {
		start := time.Now()
		valueChanges := m.updateValues()
		log.Printf("updatedValues (iteration=#%d) with %d total changes in %v", i, valueChanges, time.Since(start))
		if valueChanges == 0 {
			return nil
		}

		if err := m.Save(filePath); err != nil {
			return fmt.Errorf("Save() failed: %v", err)
		}

		start = time.Now()
		policyChanges := m.updatePolicy()
		log.Printf("updatePolicy (iteration=#%d) with %d total changes in %v", i, policyChanges, time.Since(start))
		if policyChanges == 0 {
			return nil
		}
	}
}

// Save the MDP to the filePath or returns nil if the path is empty.
func (m *MDP) Save(filePath string) error {
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
		return fmt.Errorf("decoder.Decode(previewLen): %v", err)
	}
	if err := decoder.Decode(&m.value); err != nil {
		return fmt.Errorf("decoder.Decode(value): %v", err)
	}
	continuousMoves, _ := combo4.AllContinuousMoves()
	m.nfa = combo4.NewNFA(continuousMoves)

	hasInitialVals := true
	for _, v := range m.value {
		if v == 1 {
			continue
		}
		hasInitialVals = false
		break
	}
	log.Printf("num states = %d\n", len(m.value))
	if hasInitialVals {
		m.initPolicy()
	} else {
		// Create some inital policy values.
		m.policy = make(map[GameState]combo4.State, len(m.value))
		for gState := range m.value {
			m.policy[gState] = m.nfa.NextStates(gState.State, gState.Current)[0]
		}
		m.updatePolicy()
	}
	return nil
}

// MDPPolicy contains only the information necessary to use the policy in an
// MDP.
//
// MDPPolicy is safe for concurrent use.
type MDPPolicy struct {
	policy map[GameState]combo4.State

	compressed bool
	defaultPol Policy // defaultPol is used if the policy does not contain the game state.
}

// NextState returns the next state. NextState panics if the preview is over
// length 8.
func (m *MDPPolicy) NextState(initial combo4.State, current tetris.Piece, preview []tetris.Piece, endBagUsed tetris.PieceSet) *combo4.State {
	if next, ok := m.policy[GameState{
		State:   initial,
		Current: current,
		Preview: tetris.MustSeq(preview),
		BagUsed: endBagUsed,
	}]; ok {
		copy := next
		return &copy
	}
	return m.defaultPol.NextState(initial, current, preview, endBagUsed)
}

// CompressedPolicy returns the MDP's policy in compressed form.
func (m *MDP) CompressedPolicy() *MDPPolicy {
	policy := make(map[GameState]combo4.State, len(m.policy))
	defaultPol := FromScorer(m.nfa, NewNFAScorer(m.nfa, 7))

	for gState, choice := range m.policy {
		// Only specify the choice if its not obvious.
		if choices := m.nfa.NextStates(gState.State, gState.Current); len(choices) <= 1 {
			continue
		}
		// Only specify the choice if it differs from the Scorer's policy.
		if choice == *defaultPol.NextState(gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed) {
			continue
		}
		policy[gState] = choice
	}

	log.Printf("reduced states = %d\n", len(policy))
	return &MDPPolicy{
		policy:     policy,
		defaultPol: defaultPol,
		compressed: true,
	}
}

type basicScorer struct {
	NFA *combo4.NFA
}

// Score checks how many of the known next pieces can be consumed.
func (s *basicScorer) Score(state combo4.State, next []tetris.Piece, bagUsed tetris.PieceSet) int64 {
	_, consumed := s.NFA.EndStates(combo4.NewStateSet(state), next)
	return int64(consumed)
}

// Policy returns the MDP's policy without compressing first.
func (m *MDP) Policy() Policy {
	return &MDPPolicy{
		policy:     m.policy,
		defaultPol: FromScorer(m.nfa, &basicScorer{m.nfa}),
	}
}

// GobEncode returns a Gob encoding of a MDPPolicy.
func (m *MDPPolicy) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(&m.policy); err != nil {
		return nil, fmt.Errorf("encoder.Encode(compressed): %v", err)
	}
	if err := encoder.Encode(&m.compressed); err != nil {
		return nil, fmt.Errorf("encoder.Encode(compressed): %v", err)
	}
	return buf.Bytes(), nil
}

// GobDecode decodes a Gob encoding into an MDPPolicy.
func (m *MDPPolicy) GobDecode(b []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(b) // Always returns nil.
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&m.policy); err != nil {
		return fmt.Errorf("decoder.Decode(policy): %v", err)
	}
	if err := decoder.Decode(&m.compressed); err != nil {
		return fmt.Errorf("decoder.Decode(compressed): %v", err)
	}
	continuousMoves, _ := combo4.AllContinuousMoves()
	nfa := combo4.NewNFA(continuousMoves)
	if m.compressed {
		m.defaultPol = FromScorer(nfa, NewNFAScorer(nfa, 7))
	} else {
		m.defaultPol = FromScorer(nfa, &basicScorer{nfa})
	}
	return nil
}
