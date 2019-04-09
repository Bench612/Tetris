package tetris

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSeqSetContains(t *testing.T) {
	set := NewSeqSet([]Piece{I, J, O}, []Piece{S, S, S, T, T})

	tests := []struct {
		desc string
		seq  []Piece
		want bool
	}{
		{
			desc: "Has prefix",
			seq:  []Piece{I, J, O, Z, L},
			want: true,
		},
		{
			desc: "Exact prefix match",
			seq:  []Piece{S, S, S, T, T},
			want: true,
		},
		{
			desc: "Not a match",
			seq:  []Piece{S, S, S, Z, L},
			want: false,
		},
		{
			desc: "Empty Sequence",
			seq:  nil,
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if got := set.Contains(test.seq); got != test.want {
				t.Errorf("got contains %v = %t, want %t", test.seq, got, test.want)
			}
		})
	}
}

func TestPrefixes(t *testing.T) {
	tests := []struct {
		desc string
		seqs [][]Piece
	}{
		{
			desc: "Two seqs",
			seqs: [][]Piece{
				{S, S, S, T, T},
				{I, J, O},
			},
		},
		{
			desc: "No seqs",
			seqs: nil,
		},
		{
			desc: "All seqs",
			seqs: [][]Piece{{}},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			set := NewSeqSet(test.seqs...)
			got := set.Prefixes()
			if !cmp.Equal(got, test.seqs) {
				t.Errorf("Prefixes got %v, want %v", got, test.seqs)
			}
		})
	}
}

func TestPermSize(t *testing.T) {
	// Use Fatal errors to prevent spamming.
	ps := NewPieceSet(T)
	perm := Permutations(ps)
	if got := perm.Size(7); got != 5040 {
		t.Fatalf("%s: got Size(7) = %d, want 5040", ps, got)
	}
	if got := perm.Size(0); got != 1 {
		t.Fatalf("%s: got Size(0) = %d, want 1", ps, got)
	}
	if got := perm.Size(1); got != 7-ps.Len() {
		t.Fatalf("%s: got Size(1) = %d, want %d", ps, got, 7-ps.Len())
	}
}

func TestSeqSetSize(t *testing.T) {
	tests := []struct {
		desc   string
		set    *SeqSet
		length int
		want   int
	}{
		{
			desc: "Two sequences",
			set: NewSeqSet(
				[]Piece{I, J, O},
				[]Piece{S, S, S, T, T},
			),
			length: 5,
			want:   7*7 + 1,
		},
		{
			desc: "Length 0 without [] prefix",
			set: NewSeqSet(
				[]Piece{I, J, O},
			),
			length: 0,
			want:   0,
		},
		{
			desc: "Length 0 with [] prefix",
			set: NewSeqSet(
				[]Piece{},
			),
			length: 0,
			want:   1,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := test.set.Size(test.length)
			if got != test.want {
				t.Errorf("got Size = %d, want %d", got, test.want)
			}
		})
	}
}

func TestSeqSetEquals(t *testing.T) {
	tests := []struct {
		desc  string
		seqs1 [][]Piece
		seqs2 [][]Piece
		want  bool
	}{
		{
			desc: "Prefixes should be deduped (shorter sequence first)",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, J, O},
				{I, J, O, T},
			},
			want: true,
		},
		{
			desc: "Prefixes should be deduped (longer sequence first)",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, J, O, T},
				{I, J, O},
			},
			want: true,
		},
		{
			desc:  "Exact match",
			seqs1: [][]Piece{{I, J, O}},
			seqs2: [][]Piece{{I, J, O}},
			want:  true,
		},
		{
			desc:  "Not Equal",
			seqs1: [][]Piece{{I, J, O}},
			seqs2: [][]Piece{{I, J, S}},
			want:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var (
				set1 = NewSeqSet(test.seqs1...)
				set2 = NewSeqSet(test.seqs2...)
			)

			if got := set1.Equals(set2); got != test.want {
				t.Errorf("got Equal = %t, want %t", got, test.want)
			}
		})
	}
}

func TestSeqSetIntersection(t *testing.T) {
	tests := []struct {
		desc  string
		seqs1 [][]Piece
		seqs2 [][]Piece
		want  [][]Piece
	}{
		{
			desc: "Strict subset",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, J, O, T},
			},
			want: [][]Piece{
				{I, J, O, T},
			},
		},
		{
			desc: "Strict superset",
			seqs1: [][]Piece{
				{I, J, O, T},
			},
			seqs2: [][]Piece{
				{I, J, O},
			},
			want: [][]Piece{
				{I, J, O, T},
			},
		},
		{
			desc: "Partial overlap",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, Z, O},
			},
			want: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var (
				set1 = NewSeqSet(test.seqs1...)
				set2 = NewSeqSet(test.seqs2...)
				want = NewSeqSet(test.want...)
			)

			got := set1.Intersection(set2)
			if !got.Equals(want) {
				t.Errorf("Intersection() got %v, want %v", got, want)
			}
		})
	}
}

func TestSeqSetUnion(t *testing.T) {
	tests := []struct {
		desc  string
		seqs1 [][]Piece
		seqs2 [][]Piece
		want  [][]Piece
	}{
		{
			desc: "Strict subset",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, J, O, T},
			},
			want: [][]Piece{
				{I, J, O},
			},
		},
		{
			desc: "Strict superset",
			seqs1: [][]Piece{
				{I, J, O, T},
			},
			seqs2: [][]Piece{
				{I, J, O},
			},
			want: [][]Piece{
				{I, J, O},
			},
		},
		{
			desc: "Partial overlap",
			seqs1: [][]Piece{
				{I, J, O},
			},
			seqs2: [][]Piece{
				{I, Z, O},
			},
			want: [][]Piece{
				{I, J, O},
				{I, Z, O},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var (
				set1 = NewSeqSet(test.seqs1...)
				set2 = NewSeqSet(test.seqs2...)
				want = NewSeqSet(test.want...)
			)

			got := set1.Union(set2)
			if !got.Equals(want) {
				t.Errorf("Union() got %v, want %v", got, want)
			}
		})
	}
}

func TestPrependedSeqSets(t *testing.T) {
	initial := NewSeqSet([]Piece{I, J, O}, []Piece{S, Z, L})
	want := NewSeqSet([]Piece{S, I, J, O}, []Piece{S, S, Z, L})

	var prefixToSet [8]*SeqSet
	prefixToSet[S] = initial
	got := PrependedSeqSets(prefixToSet)
	if !got.Equals(want) {
		t.Errorf("PrependedSeqSets got %v, want %v", got, want)
	}
}
