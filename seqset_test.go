package tetris

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSeqSetContains(t *testing.T) {
	set := new(SeqSet)
	set.AddPrefix([]Piece{I, J, O})
	set.AddPrefix([]Piece{S, S, S, T, T})

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
			set := new(SeqSet)
			for _, seq := range test.seqs {
				set.AddPrefix(seq)
			}
			got := set.Prefixes()
			if !cmp.Equal(got, test.seqs) {
				t.Errorf("Prefixes got %v, want %v", got, test.seqs)
			}
		})
	}
}

func TestSeqSetEncodeDecode(t *testing.T) {
	seqs := new(SeqSet)
	seqs.AddPrefix([]Piece{I, J, O})
	bytes1, _ := seqs.GobEncode()

	got := &SeqSet{}
	if err := got.GobDecode(bytes1); err != nil {
		t.Fatalf("GobDecode failed: %v", err)
	}
	if !got.Equals(seqs) {
		t.Errorf("Encode->Decode does not equal original")
	}
	if !cmp.Equal(got.Prefixes(), seqs.Prefixes()) {
		t.Errorf("Encode->Decode prefixes does not equal original got %v, want %v", got.Prefixes(), seqs.Prefixes())
	}
	bytes2, _ := got.GobEncode()
	if !bytes.Equal(bytes1, bytes2) {
		diff := cmp.Diff(bytes1, bytes2)
		t.Errorf("Second GobEncoding is not equal to first (-bytes1 +bytes2)\n:%s", diff)
	}

}

func TestSeqSetSize(t *testing.T) {
	tests := []struct {
		desc   string
		seqs   [][]Piece
		length int
		want   int
	}{
		{
			desc: "Two sequences",
			seqs: [][]Piece{
				{I, J, O},
				{S, S, S, T, T},
			},
			length: 5,
			want:   7*7 + 1,
		},
		{
			desc:   "Length 0 without [] prefix",
			seqs:   [][]Piece{{I, J, O}},
			length: 0,
			want:   0,
		},
		{
			desc:   "Length 0 with [] prefix",
			seqs:   [][]Piece{{}},
			length: 0,
			want:   1,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			set := new(SeqSet)
			for _, seq := range test.seqs {
				set.AddPrefix(seq)
			}
			got := set.Size(test.length)
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
			set1 := new(SeqSet)
			for _, seq := range test.seqs1 {
				set1.AddPrefix(seq)
			}

			set2 := new(SeqSet)
			for _, seq := range test.seqs2 {
				set2.AddPrefix(seq)
			}

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
			set1 := new(SeqSet)
			for _, seq := range test.seqs1 {
				set1.AddPrefix(seq)
			}

			set2 := new(SeqSet)
			for _, seq := range test.seqs2 {
				set2.AddPrefix(seq)
			}

			want := new(SeqSet)
			for _, seq := range test.want {
				want.AddPrefix(seq)
			}

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
			set1 := new(SeqSet)
			for _, seq := range test.seqs1 {
				set1.AddPrefix(seq)
			}

			set2 := new(SeqSet)
			for _, seq := range test.seqs2 {
				set2.AddPrefix(seq)
			}

			want := new(SeqSet)
			for _, seq := range test.want {
				want.AddPrefix(seq)
			}

			got := set1.Union(set2)
			if !got.Equals(want) {
				t.Errorf("Union() got %v, want %v", got, want)
			}
		})
	}
}
