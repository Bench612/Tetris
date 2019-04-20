package bot

import (
	"testing"
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
	mdp, err := NewMDP(1, -1)
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
	mdp, err := NewMDP(1, -1)
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
