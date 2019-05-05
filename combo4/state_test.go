package combo4

import (
	"testing"
	"tetris"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkNFA7(b *testing.B) {
	benchmarkNFA(b, 7)
}
func BenchmarkNFA20(b *testing.B) {
	benchmarkNFA(b, 20)
}

func BenchmarkNFA400(b *testing.B) {
	benchmarkNFA(b, 400)
}

func BenchmarkNFA700(b *testing.B) {
	benchmarkNFA(b, 700)
}

func benchmarkNFA(b *testing.B, sequenceLen int) {
	moves, _ := AllContinuousMoves()
	nfa := NewNFA(moves)

	inputs := make([][]tetris.Piece, 0, b.N)
	for n := 0; n < b.N; n++ {
		inputs = append(inputs, tetris.RandPieces(sequenceLen))
	}

	b.ResetTimer()
	var completed int
	for n := 0; n < b.N; n++ {
		_, consumed := nfa.EndStates(NewStateSet(State{Field: RightI}), inputs[n])
		if consumed == len(inputs[n]) {
			completed++
		}
	}
	b.Logf("Number of end states with possibilities %.3f%% of %d tries", float64(completed)/float64(b.N), b.N)
}

func TestEndStates(t *testing.T) {
	moves, _ := AllContinuousMoves()
	nfa := NewNFA(moves)

	const X, o = true, false

	tests := []struct {
		desc          string
		initState     State
		pieces        []tetris.Piece
		wantEndStates StateSet
		wantConsumed  int
	}{
		{
			desc:      "Should consume all",
			initState: State{Field: LeftI},
			pieces:    []tetris.Piece{tetris.S, tetris.O, tetris.L},
			wantEndStates: NewStateSet(
				State{
					Field:          NewField4x4([][4]bool{{X, X, X, o}}),
					Hold:           tetris.L,
					SwapRestricted: true,
				},
				State{
					Field: NewField4x4([][4]bool{
						{X, o, o, o},
						{X, o, X, o},
					}),
					Hold: tetris.O,
				},
				State{
					Field: NewField4x4([][4]bool{
						{o, o, X, X},
						{o, o, o, X},
					}),
				},
			),
			wantConsumed: 3,
		},
		{
			desc:      "Should leave one unconsumed",
			initState: State{Field: LeftI},
			pieces:    []tetris.Piece{tetris.J, tetris.O, tetris.S},
			wantEndStates: NewStateSet(
				State{
					Field:          NewField4x4([][4]bool{{o, X, X, X}}),
					Hold:           tetris.O,
					SwapRestricted: true,
				},
			),
			wantConsumed: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotEndStates, gotConsumed := nfa.EndStates(NewStateSet(test.initState), test.pieces)
			if diff := cmp.Diff(map[State]bool(test.wantEndStates), map[State]bool(gotEndStates)); diff != "" {
				t.Errorf("end states mismatch(-want +got):\n%s", diff)
			}
			if test.wantConsumed != gotConsumed {
				t.Errorf("consumed pieces = %d, want %d", gotConsumed, test.wantConsumed)
			}
		})
	}
}

func TestStateSetEqual(t *testing.T) {
	tests := []struct {
		desc string
		a, b StateSet
		want bool
	}{
		{
			desc: "Same sets",
			a:    NewStateSet(State{Field: LeftI}),
			b:    NewStateSet(State{Field: LeftI}),
			want: true,
		},
		{
			desc: "Different sets",
			a:    NewStateSet(State{Field: LeftI}),
			b:    NewStateSet(State{Field: RightI}),
			want: false,
		},
		{
			desc: "Sets with diffferent lengths",
			a:    NewStateSet(State{Field: LeftI}),
			b:    NewStateSet(State{Field: RightI}, State{Field: LeftI}),
			want: false,
		},
		{
			desc: "Nil StateSet and empty StateSet",
			a:    nil,
			b:    NewStateSet(),
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if got := test.a.Equals(test.b); got != test.want {
				t.Errorf("a.Equals(b)=%t, want %t", got, test.want)
			}
			if got := test.b.Equals(test.a); got != test.want {
				t.Errorf("b.Equals(a)=%t, want %t", got, test.want)
			}
		})
	}
}

func TestStateSetSlice(t *testing.T) {
	states := []State{{Field: LeftI}}
	set := NewStateSet(states...)
	if got := set.Slice(); !cmp.Equal(got, states) {
		t.Errorf("Slice() got %v, want %v", got, states)
	}
}

func TestNextStates(t *testing.T) {
	startState := State{Field: LeftI}
	piece := tetris.L

	want := []State{{Field: LeftI, Hold: tetris.L}}

	nfa := new(NFA)
	nfa.trans[piece] = map[State][]State{
		startState: want,
	}

	if got := nfa.NextStates(startState, piece); !cmp.Equal(got, want) {
		t.Errorf("NextStates() got %v, want %v", got, want)
	}
}
