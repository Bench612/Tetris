package bot

import "testing"

func BenchmarkNewMDP3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		if _, err := NewMDP(3); err != nil {
			b.Fatalf("NewMDP failed: %v", err)
		}
	}
}
