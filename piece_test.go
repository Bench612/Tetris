package tetris

import (
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func BenchmarkPieceSetToSlice(b *testing.B) {
	inputs := make([]PieceSet, b.N)
	for n := 0; n < b.N; n++ {
		inputs[n] = PieceSet(rand.Int31())
	}
	b.ResetTimer()

	totalLengths := 0
	for n := 0; n < b.N; n++ {
		totalLengths += len(inputs[n].ToSlice())
	}
	b.Logf("Average length is %.2f in %d tries", float64(totalLengths)/float64(b.N), b.N)
}

func BenchmarkPieceSeq(b *testing.B) {
	inputs := make([][]Piece, b.N)
	for n := 0; n < b.N; n++ {
		inputs[n] = RandPieces(7)
	}
	b.ResetTimer()

	totalLengths := 0
	for n := 0; n < b.N; n++ {
		seq, err := NewPieceSeq(inputs[n])
		if err != nil {
			b.Errorf("NewPieceSeq failed: %v", err)
		}
		totalLengths += len(seq.ToSlice())
	}
	b.Logf("Average length is %.2f in %d tries", float64(totalLengths)/float64(b.N), b.N)
}

func TestToSlice(t *testing.T) {
	tests := []struct {
		desc  string
		input []Piece
		want  []Piece
	}{
		{
			desc: "No pieces",
		},
		{
			desc:  "EmptyPiece and O should not include the EmptyPiece",
			input: []Piece{EmptyPiece, O},
			want:  []Piece{O},
		},
		{
			desc:  "3 Pieces",
			input: []Piece{S, O, I},
			want:  []Piece{S, O, I},
		},
		{
			desc:  "Duplicate Piece",
			input: []Piece{I, I, S},
			want:  []Piece{S, I},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ps := NewPieceSet(test.input...)
			got := ps.ToSlice()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("ToSlice() mismatch(-want +got):\n%s", diff)
			}
		})
	}
}

func TestLen(t *testing.T) {
	tests := []struct {
		desc   string
		pieces []Piece
		want   int
	}{
		{
			desc: "No pieces",
			want: 0,
		},
		{
			desc:   "Beginning and end of range",
			pieces: []Piece{EmptyPiece, I},
			want:   1,
		},
		{
			desc:   "3 Pieces",
			pieces: []Piece{I, S, O},
			want:   3,
		},
		{
			desc:   "Duplicate Piece",
			pieces: []Piece{I, I, S},
			want:   2,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			ps := NewPieceSet(test.pieces...)
			if got := ps.Len(); got != test.want {
				t.Errorf("Len() got %d, want %d", got, test.want)
			}
		})
	}
}

func TestInverted(t *testing.T) {
	tests := []struct {
		desc     string
		pieceSet PieceSet
		want     PieceSet
	}{
		{
			desc:     "Three pieces",
			pieceSet: NewPieceSet(I, O, S),
			want:     NewPieceSet(T, L, J, Z),
		},
		{
			desc: "No Pieces",
			want: NewPieceSet(NonemptyPieces[:]...),
		},
		{
			desc:     "All Pieces",
			pieceSet: NewPieceSet(NonemptyPieces[:]...),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if got := test.pieceSet.Inverted(); got != test.want {
				t.Errorf("Inverted() got %v, want %v", got, test.want)
			}
		})
	}
}

func TestRandPieces(t *testing.T) {
	r7 := RandPieces(7)
	if set := NewPieceSet(r7...); set != NewPieceSet(NonemptyPieces[:]...) {
		t.Errorf("RandPiece(7) does not contain all pieces, got %v", r7)
	}

	r10 := RandPieces(10)
	if len(r10) != 10 {
		t.Errorf("RandPiece(10) got len=%d, want 10", len(r10))
	}
}

func TestAddPiece(t *testing.T) {
	var empty PieceSet

	got := empty.Add(S)
	want := NewPieceSet(S)
	if got != want {
		t.Errorf("empty.Add(S) got %v, want %v", got, want)
	}
}

func TestUnion(t *testing.T) {
	st := NewPieceSet(S, T)
	tj := NewPieceSet(T, J)

	want := NewPieceSet(S, T, J)
	if got := st.Union(tj); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNewPieceSeq(t *testing.T) {
	tests := []struct {
		desc   string
		pieces []Piece
	}{
		{
			desc:   "Three pieces",
			pieces: []Piece{I, L, O},
		},
		{
			desc:   "7  pieces",
			pieces: []Piece{I, L, O, S, J, S, I},
		},
		{
			desc:   "Include empty piece",
			pieces: []Piece{I, EmptyPiece, O, EmptyPiece},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			seq, err := NewPieceSeq(test.pieces)
			if err != nil {
				t.Fatalf("NewPieceSeq failed: %v", err)
			}
			got := seq.ToSlice()
			if diff := cmp.Diff(test.pieces, got); diff != "" {
				t.Errorf("ToSlice() mismatch(-want +got):\n%s", diff)
			}
		})
	}
}

func TestPieceSeqAppend(t *testing.T) {
	var seq PieceSeq
	want := []Piece{S, EmptyPiece, O}
	for _, p := range want {
		var err error
		seq, err = seq.Append(p)
		if err != nil {
			t.Fatalf("Append(%s) failed", p)
		}
	}
	if diff := cmp.Diff(MustPieceSeq(want).String(), seq.String()); diff != "" {
		t.Errorf("sequence mismatch(-want +got):\n%s", diff)
	}
}

func TestPieceSeqEncodeDecode(t *testing.T) {
	seq := MustPieceSeq(NonemptyPieces[:])
	bytes, err := seq.GobEncode()
	if err != nil {
		t.Fatalf("GobEncode failed: %v", err)
	}

	got := &PieceSeq{}
	if err := got.GobDecode(bytes); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	if diff := cmp.Diff(seq.ToSlice(), got.ToSlice()); diff != "" {
		t.Errorf("mismatch after encoding + decoding (-want +got):\n%s", diff)
	}
}
