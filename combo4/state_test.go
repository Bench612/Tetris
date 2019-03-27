package combo4

import (
	"testing"
	"tetris"

	"github.com/google/go-cmp/cmp"
)

// Previous 188516 ns
func BenchmarkNFA50(b *testing.B) {
	nfa := NewNFA()

	inputs := make([][]tetris.Piece, 0, b.N)
	for n := 0; n < b.N; n++ {
		inputs = append(inputs, tetris.RandPieces(50))
	}

	initialStates := NewStateSet(State{Field: RightI})

	b.ResetTimer()
	var completed int
	for n := 0; n < b.N; n++ {
		_, unconsumed := nfa.EndStates(initialStates, inputs[n]...)
		if len(unconsumed) == 0 {
			completed++
		}
	}
	b.Logf("Number of end states with possibilities %.3f%% of %d tries", float64(completed)/float64(b.N), b.N)
}

func TestEndStates(t *testing.T) {
	nfa := NewNFA()
	const X, o = true, false

	tests := []struct {
		desc           string
		initStates     StateSet
		pieces         []tetris.Piece
		wantEndStates  StateSet
		wantUnconsumed []tetris.Piece
	}{
		{
			desc:       "Should consume all",
			initStates: NewStateSet(State{Field: LeftI}),
			pieces:     []tetris.Piece{tetris.S, tetris.O, tetris.L},
			wantEndStates: NewStateSet(
				State{
					Field: NewField4x4([][4]bool{{X, X, X, o}}),
					Hold:  tetris.L,
				},
				State{
					Field: NewField4x4([][4]bool{
						{X, o, o, o},
						{X, o, X, o},
					}),
					Hold: tetris.O,
				},
			),
		},
		{
			desc:       "Should leave one unconsumed",
			initStates: NewStateSet(State{Field: LeftI}),
			pieces:     []tetris.Piece{tetris.J, tetris.O, tetris.S},
			wantEndStates: NewStateSet(
				State{
					Field: NewField4x4([][4]bool{{o, X, X, X}}),
					Hold:  tetris.O,
				},
			),
			wantUnconsumed: []tetris.Piece{tetris.S},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotEndStates, gotUnconsumed := nfa.EndStates(test.initStates, test.pieces...)
			if diff := cmp.Diff(map[State]bool(test.wantEndStates), map[State]bool(gotEndStates)); diff != "" {
				t.Errorf("end states mismatch(-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantUnconsumed, gotUnconsumed); diff != "" {
				t.Errorf("unconsumed pieces mismatch(-want +got):\n%s", diff)
			}
		})
	}
}
