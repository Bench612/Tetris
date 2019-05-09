package tetris

import "testing"

func TestActionMirror(t *testing.T) {
	mirrorTaken := make(map[Action]Action)
	for a := Action(0); a < actionLimit; a++ {
		mirror := a.Mirror()
		if other, ok := mirrorTaken[mirror]; ok {
			t.Errorf("Action %v has the same Mirror as %v", other, a)
		}
		mirrorTaken[mirror] = a

		if got := mirror.Mirror(); got != a {
			t.Errorf("%v.Mirror().Mirror() got %v, want %v", a, got, a)
		}
	}
}

func TestActionString(t *testing.T) {
	counts := make(map[string]int)
	for a := Action(0); a < actionLimit; a++ {
		s := a.String()
		counts[s]++
		if counts[s] > 1 {
			t.Errorf("String %q is repeated for %d", s, int(a))
		}
		if s == "Unknown" {
			t.Errorf("Action %d has no mapped string", int(a))
		}
	}
}
