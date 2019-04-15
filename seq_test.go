package tetris

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewSeq(t *testing.T) {
	tests := []struct {
		desc   string
		pieces []Piece
	}{
		{
			desc:   "3 pieces",
			pieces: []Piece{I, L, O},
		},
		{
			desc:   "8  pieces",
			pieces: []Piece{I, L, O, S, J, S, I, I},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			seq, err := NewSeq(test.pieces)
			if err != nil {
				t.Fatalf("NewSeq failed: %v", err)
			}
			got := seq.Slice()
			if diff := cmp.Diff(test.pieces, got); diff != "" {
				t.Errorf("Slice() mismatch(-want +got):\n%s", diff)
			}
		})
	}
}

func TestSetIndex(t *testing.T) {
	tests := []struct {
		desc   string
		pieces []Piece
		set    Piece
		setIdx int
		want   []Piece
	}{
		{
			desc:   "Append to end",
			pieces: []Piece{I, L, O},
			set:    J,
			setIdx: 3,
			want:   []Piece{I, L, O, J},
		},
		{
			desc:   "Set beginning",
			pieces: []Piece{I, L, O},
			set:    J,
			setIdx: 0,
			want:   []Piece{J, L, O},
		},
		{
			desc:   "Set end",
			pieces: []Piece{I, L, O},
			set:    J,
			setIdx: 2,
			want:   []Piece{I, L, J},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			seq, err := NewSeq(test.pieces)
			if err != nil {
				t.Fatalf("NewSeq failed: %v", err)
			}
			got := seq.SetIndex(test.setIdx, test.set)
			if got != MustSeq(test.want) {
				diff := cmp.Diff(test.want, got.Slice())
				t.Errorf("mismatch(-want +got):\n%s", diff)
			}
		})
	}
}

func TestRemoveFirst(t *testing.T) {
	tests := []struct {
		desc   string
		pieces []Piece
		set    Piece
		setIdx int
		want   []Piece
	}{
		{
			desc:   "3 pieces",
			pieces: []Piece{I, L, O},
			want:   []Piece{L, O},
		},
		{
			desc:   "1 piece",
			pieces: []Piece{L},
			want:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			seq, err := NewSeq(test.pieces)
			if err != nil {
				t.Fatalf("NewSeq failed: %v", err)
			}
			got := seq.RemoveFirst().Slice()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Slice() mismatch(-want +got):\n%s", diff)
			}
		})
	}
}
