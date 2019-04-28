package bot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkNewMDP3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := NewMDP(3, -1); err != nil {
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
	mdp, err := NewMDP(3, -1)
	if err != nil {
		b.Fatalf("NewMDP: %v", err)
	}
	mdp.updateValues()
}

func benchmarkMDPUpdate(b *testing.B, previewLen int) {
	for n := 0; n < b.N; n++ {
		mdp, err := NewMDP(previewLen, -1)
		if err != nil {
			b.Fatalf("NewMDP: %v", err)
		}
		mdp.Update("")

		var maxVal int
		for _, v := range mdp.value {
			if v > maxVal {
				maxVal = v
			}
		}
		b.Logf("maxVal=%d", maxVal)
	}
}

func TestMDPUpdate(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1, -1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.Update("")
	var maxVal int
	for _, v := range mdp.value {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal != 44 {
		t.Errorf("got maximum value = %d, want 44", maxVal)
	}
}

func TestMDPUpdateValues(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1, 10)
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

func TestMDPUpdatePolicy(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1, 10)
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

	mdp, err := NewMDP(1, 4)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}

	t.Run("without update", func(t *testing.T) { testMdpGobHelper(t, mdp) })

	mdp.updateValues()
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
	if decoding.maxValue != mdp.maxValue {
		t.Errorf("got maxValue=%d after decoding, want %d", decoding.maxValue, mdp.maxValue)
	}
}
