package tetris

import "testing"

func TestActionMirror(t *testing.T) {
	actions := []Action{NoAction, Hold, Left, Right, RotateCW, RotateCCW, SoftDrop, HardDrop}

	mirrorTaken := make(map[Action]Action)
	for _, a := range actions {
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
