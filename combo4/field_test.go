package combo4

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestField4x4(t *testing.T) {
	tests := []struct {
		desc            string
		input           [][4]bool
		wantString      string
		wantNumOccupied int
	}{
		{
			desc: "Two rows should place rows on the bottom",
			input: [][4]bool{
				{true, false, false, false},
				{true, true, false, false},
			},
			wantString:      "□___\n□□__\n",
			wantNumOccupied: 3,
		},
		{
			desc: "Six rows should use the bottom 4",
			input: [][4]bool{
				{true, true, true, true},
				{true, true, true, true},
				{false, false, false, false},
				{false, false, false, false},
				{true, false, false, false},
				{true, true, false, false},
			},
			wantString:      "□___\n□□__\n",
			wantNumOccupied: 3,
		},
		{
			desc:            "No rows",
			input:           [][4]bool{},
			wantString:      "",
			wantNumOccupied: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			f4x4 := NewField4x4(test.input)
			if diff := cmp.Diff(test.wantString, f4x4.String()); diff != "" {
				t.Errorf("String() mismatch(-want +got):\n%s", diff)
			}
			if got := f4x4.NumOccupied(); got != test.wantNumOccupied {
				t.Errorf("NumOccupied() got %d, want %d", got, test.wantNumOccupied)
			}
		})
	}
}

func TestField4x4Mirror(t *testing.T) {
	const X, o = true, false

	tests := []struct {
		desc  string
		input Field4x4
		want  Field4x4
	}{
		{
			desc: "Two rows",
			input: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			want: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, X, X},
			}),
		},
		{
			desc: "One rows",
			input: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			want: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := test.input.Mirror()
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("Mirror got mismatch(-want +got):\n%s", diff)
			}
		})
	}
}

func TestFieldConsants(t *testing.T) {
	const X, o = true, false

	tests := []struct {
		desc  string
		input Field4x4
		want  Field4x4
	}{
		{
			desc:  "LeftI",
			input: LeftI,
			want:  NewField4x4([][4]bool{{true, true, true, false}}),
		},
		{
			desc:  "RightI",
			input: RightI,
			want:  NewField4x4([][4]bool{{false, true, true, true}}),
		},
		{
			desc:  "LeftZ",
			input: LeftZ,
			want:  NewField4x4([][4]bool{{true, false, false, false}, {true, true, false, false}}),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if diff := cmp.Diff(test.input, test.want); diff != "" {
				t.Errorf("mismatch(-want +got):\n%s", diff)
			}
		})
	}
}
