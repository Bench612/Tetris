package bot

import (
	"testing"
)

func BenchmarkNewMDP3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := NewMDP(3); err != nil {
			b.Fatalf("NewMDP failed: %v", err)
		}
	}
}

func BenchmarkMDPUpdate(b *testing.B) {
	for n := 0; n < b.N; n++ {
		mdp, err := NewMDP(1)
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

func TestUpdateValues(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.UpdateValues()
	// Trying to update the values again should show no change since the
	// first one should iterate until equilibrium.
	if mdp.UpdateValues() {
		t.Errorf("2nd UpdateValues call had changes")
	}
}

func TestUpdatePolicy(t *testing.T) {
	t.Parallel()
	mdp, err := NewMDP(1)
	if err != nil {
		t.Fatalf("NewMDP: %v", err)
	}
	mdp.UpdateValues()
	for mdp.UpdatePolicy() {
	}
	if mdp.UpdatePolicy() {
		t.Errorf("UpdatePolicy call after no changes had changes")
	}
}
