package bot

import (
	"math"
	"math/rand"
	"testing"
	"tetris"
	"tetris/combo4"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkNewMDP3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := NewMDP(3); err != nil {
			b.Fatalf("NewMDP failed: %v", err)
		}
	}
}

func BenchmarkMDP1Update(b *testing.B) {
	benchmarkMDPUpdate(b, 1)
}

func BenchmarkMDP2Update(b *testing.B) {
	benchmarkMDPUpdate(b, 2)
}

func BenchmarkMDP3UpdateValues(b *testing.B) {
	mdp, err := NewMDP(3)
	if err != nil {
		b.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()
}

func benchmarkMDPUpdate(b *testing.B, previewLen int) {
	for n := 0; n < b.N; n++ {
		mdp, err := NewMDP(previewLen)
		if err != nil {
			b.Fatalf("NewMDP: %v", err)
		}
		mdp.Update("")

		var maxVal float64
		for _, v := range mdp.value {
			if v > maxVal {
				maxVal = v
			}
		}
		b.Logf("maxVal=%.2f", maxVal)
	}
}

func TestMDPUpdateValues(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()
	// Trying to update the values again should show no change since the
	// first one should iterate until equilibrium.
	if mdp.updateValues() != 0 {
		t.Errorf("2nd UpdateValues call had changes")
	}
}

func TestCompressedPolicy(t *testing.T) {
	t.Parallel()

	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()
	mdp.updatePolicy()

	compressed := mdp.CompressedPolicy()
	policy := mdp.Policy()

	// Verify the choice is the same for each "stable" GameState.
	for gState := range mdp.value {
		got := compressed.NextState(gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed)
		want := policy.NextState(gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed)
		if !cmp.Equal(got, want) {
			t.Fatalf("Compressed policy differs for state %v", gState)
		}
	}
}

// This test is technically flaky but has a low failure rate because it
// takes a lot of samples.
func TestMDPExpectedValue(t *testing.T) {
	t.Parallel()

	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()

	// Check that the expected value of a GameState is accurate by doing
	// some sampling.
	policy := mdp.Policy()
	gState := GameState{
		State: combo4.State{
			Hold: tetris.J,
			Field: combo4.NewField4x4([][4]bool{
				{true, false, false, false},
				{true, true, false, false},
			}),
		},
		Current: tetris.S,
		Preview: tetris.MustSeq([]tetris.Piece{tetris.O}),
		BagUsed: tetris.NewPieceSet(tetris.O, tetris.S),
	}

	const numTrials = 20 * 1000
	var sampleValue float64
	for trial := 0; trial < numTrials; trial++ {
		inputCh := make(chan tetris.Piece, 7)
		outputCh := ResumeGame(policy, gState.State, gState.Current, gState.Preview.Slice(), gState.BagUsed, inputCh)

		// Populate the inputCh with some initial values.
		initial := gState.BagUsed.Inverted().Slice()
		rand.Shuffle(len(initial), func(i, j int) { initial[i], initial[j] = initial[j], initial[i] })
		for _, p := range initial {
			inputCh <- p
		}

		var count int
	OuterLoop:
		for {
			next := tetris.RandPieces(7)
			for _, p := range next {
				if <-outputCh == nil {
					break OuterLoop
				}
				count++
				inputCh <- p
			}
		}
		sampleValue += float64(count) / numTrials
	}

	if got := mdp.ExpectedValue(gState); math.Abs(got-sampleValue) > 1 {
		t.Errorf("got ExpectedValue=%.2f, want %.2f", got, sampleValue)
	}
}

func TestMDPUpdatePolicy(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()
	for mdp.updatePolicy() != 0 {
	}
	if mdp.updatePolicy() != 0 {
		t.Errorf("updatePolicy call after no changes had changes")
	}
}

func TestMDPGob(t *testing.T) {
	t.Parallel()

	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}

	t.Run("without update", func(t *testing.T) { testMdpGobHelper(t, mdp) })

	mdp.Update("")
	t.Run("with update", func(t *testing.T) { testMdpGobHelper(t, mdp) })
}

func testMdpGobHelper(t *testing.T, mdp *MDP) {
	encoding1, err := mdp.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode: %v", err)
	}

	decoding := new(MDP)
	if err := decoding.GobDecode(encoding1); err != nil {
		t.Fatalf("GobDecode: %v", err)
	}

	if diff := cmp.Diff(decoding.value, mdp.value); diff != "" {
		t.Errorf("value map differs after decoding: (-want +got)\n:%v", diff)
	}
	if decoding.previewLen != 1 {
		t.Errorf("got previewLen=%d after decoding, want 1", decoding.previewLen)
	}
}

func TestMDPPolicyGob(t *testing.T) {
	t.Parallel()

	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}

	policy := (mdp.Policy()).(*MDPPolicy)

	encoding1, err := policy.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode: %v", err)
	}

	decoding := new(MDPPolicy)
	if err := decoding.GobDecode(encoding1); err != nil {
		t.Fatalf("GobDecode: %v", err)
	}

	if diff := cmp.Diff(decoding.policy, policy.policy); diff != "" {
		t.Errorf("value map differs after decoding: (-want +got)\n:%v", diff)
	}
}
