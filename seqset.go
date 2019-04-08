package tetris

import (
	"bytes"
	"fmt"
)

func init() {
	// Generate the permutations.
	for _, bag := range AllPieceSets() {
		bagIdx := bag
		permutations[bagIdx].permBag = &bagIdx

		// Full bag is equivalent to empty bag.
		if bag.Len() == 7 {
			bag = 0
		}
		for _, p := range NonemptyPieces {
			if !bag.Contains(p) {
				newBag := bag.Add(p)
				permutations[bagIdx].subSeqSets[p-1] = &permutations[newBag]
			}
		}
	}
}

// SeqSet represents a set of sequences.
//
// A SeqSet is defined by prefixes. For example, given a SeqSet with the
// prefix [T O I], the SeqSet contains [T O I], [T O I T], [T O I Z Z], etc.
// This can be useful for storing sequences that fail since you know that all
// sequences with that prefix will also fail.
//
// The nil pointer is usable.
type SeqSet struct {
	hasAllSeq  bool       // Whether all sequences are contained.
	subSeqSets [7]*SeqSet // "map" from (Piece-1) to a SeqSet.

	// This is a special value that is only used for the global var
	// "permutations". See the comment above the "permutations" variable.
	permBag *PieceSet
}

// ContainsAllSeqSet is a SeqSet that contains all sequences.
var ContainsAllSeqSet = &SeqSet{hasAllSeq: true}

// permutations is a collection of special SeqSets from PieceSet->SeqSet.
// Each SeqSet contains all possible permutations from the given bag state.
// These SeqSets are special because
// a) Only these SeqSets, SeqSets containing these, and copies can have cycles.
// b) These SeqSets contain all Sequences along the traversal path.
var permutations [255]SeqSet

// Permutations returns a SeqSet that contains all sequences starting from the
// given bag state assuming a 7 bag randomizer.
func Permutations(bagUsed PieceSet) *SeqSet {
	return &permutations[bagUsed]
}

// AddPrefix adds the prefix to the SeqSet.
func (s *SeqSet) AddPrefix(prefix []Piece) {
	if s.hasAllSeq {
		return
	}
	if s.permBag != nil {
		panic("SeqSets from the Permutations func cannot be modified")
	}
	if len(prefix) == 0 {
		s.hasAllSeq = true
		// Zero out all the sub sequences sets which are now redundant.
		for i := 0; i < len(s.subSeqSets); i++ {
			s.subSeqSets[i] = nil
		}
		return
	}
	if prefix[0] == EmptyPiece {
		panic("cannot add prefixes with EmptyPieces in them")
	}
	if len(prefix) == 1 {
		s.subSeqSets[prefix[0]-1] = ContainsAllSeqSet
		return
	}

	next := s.subSeqSets[prefix[0]-1]
	if next == nil {
		next = new(SeqSet)
		s.subSeqSets[prefix[0]-1] = next
	}
	next.AddPrefix(prefix[1:])
}

// Contains returns if the sequence is contained in the SeqSet.
// Contains panics if the sequence contains an EmptyPiece.
func (s *SeqSet) Contains(sequence []Piece) bool {
	if s == nil {
		return false
	}
	if s.hasAllSeq {
		return true
	}
	if len(sequence) == 0 {
		// Permutations contain all sequences that dont lead to nil.
		return s.permBag != nil
	}
	sub := s.subSeqSets[sequence[0]-1]
	return sub.Contains(sequence[1:])
}

// Prefixes returns the prefixes contained in this SeqSet.
func (s *SeqSet) Prefixes() [][]Piece {
	all := s.reversedPrefixes(0)
	for _, p := range all {
		// Reverse each of the reversed prefixes.
		for i := 0; i < len(p)/2; i++ {
			opp := len(p) - 1 - i
			p[i], p[opp] = p[opp], p[i]
		}
	}
	return all
}

// reversedPrefixes returns all prefixes in reverse. This is more efficient
// because slices are better to append to instead of prepend.
func (s *SeqSet) reversedPrefixes(depth int) [][]Piece {
	if s == nil || s.permBag != nil {
		return nil
	}
	if s.hasAllSeq {
		return [][]Piece{
			make([]Piece, 0, depth),
		}
	}
	var all [][]Piece
	for idx, sub := range s.subSeqSets {
		piece := Piece(idx + 1)
		for _, subPrefix := range sub.reversedPrefixes(depth + 1) {
			prefix := append(subPrefix, piece)
			all = append(all, prefix)
		}
	}
	return all
}

func (s *SeqSet) String() string {
	if s.hasAllSeq {
		return "{prefixes=all}"
	}
	return fmt.Sprintf("{prefixes=%v}", s.Prefixes())
}

// Union returns the union of this SeqSet and another.
//
// WARNING: The input and output SeqSets may have the same underlying data and
// modifications to either have undefined behavior.
func (s *SeqSet) Union(other *SeqSet) *SeqSet {
	if s == nil {
		return other
	}
	if other == nil {
		return s
	}
	if s.hasAllSeq || other.hasAllSeq {
		return ContainsAllSeqSet
	}
	union := &SeqSet{}
	for i := range union.subSeqSets {
		union.subSeqSets[i] = s.subSeqSets[i].Union(other.subSeqSets[i])
	}
	return union
}

// Intersection returns the intersection of this SeqSet and another.
//
// WARNING: The input and output SeqSets may have the same underlying data and
// modifications to either have undefined behavior.
func (s *SeqSet) Intersection(other *SeqSet) *SeqSet {
	if s == nil || other == nil {
		return nil
	}
	if s.hasAllSeq {
		return other
	}
	if other.hasAllSeq {
		return s
	}
	intersect := &SeqSet{}
	var hasSubSeq bool
	for i := range intersect.subSeqSets {
		subInter := s.subSeqSets[i].Intersection(other.subSeqSets[i])
		if subInter != nil {
			intersect.subSeqSets[i] = subInter
			hasSubSeq = true
		}
	}
	if hasSubSeq {
		return intersect
	}
	return nil
}

// Size returns the total number of sequences of a given length in the SeqSet.
func (s *SeqSet) Size(length int) int {
	if s == nil {
		return 0
	}
	if s.permBag != nil {
		// Calculate the number of sequences by the choices at each step
		// assuming a 7 bag randomizer.
		choices := 7 - (*s.permBag).Len()
		if choices == 0 {
			choices = 7
		}

		prod := 1
		for i := 0; i < length; i++ {
			prod *= choices

			choices--
			if choices == 0 {
				choices = 7
			}
		}
		return prod
	}
	if length < 0 {
		return 0
	}
	if s.hasAllSeq {
		prod := 1
		for i := 0; i < length; i++ {
			prod *= 7
		}
		return prod
	}
	sum := 0
	for _, sub := range s.subSeqSets {
		sum += sub.Size(length - 1)
	}
	return sum
}

// GobEncode returns a bytes representation of the SeqSet.
// GobEncode always returns a nil error.
func (s *SeqSet) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if s == nil {
		new(SeqSet).encodeToBuffer(buf)
	} else {
		s.encodeToBuffer(buf)
	}
	return buf.Bytes(), nil
}

func (s SeqSet) encodeToBuffer(buf *bytes.Buffer) {
	if s.hasAllSeq {
		buf.WriteByte(128)
		return
	}

	// If the SeqSet is a permutation.
	if s.permBag != nil {
		buf.WriteByte(255)
		buf.WriteByte(byte(*s.permBag))
		return
	}

	// Capture which indices are null.
	var b uint8
	for idx, sub := range s.subSeqSets {
		if sub != nil {
			b |= 1 << uint(idx)
		}
	}
	buf.WriteByte(b) // Always returns nil
	for _, sub := range s.subSeqSets {
		if sub != nil {
			sub.encodeToBuffer(buf)
		}
	}
}

// GobDecode decodes a bytes representation of SeqSet into the reciever.
func (s *SeqSet) GobDecode(data []byte) error {
	buf := new(bytes.Buffer)
	buf.Write(data) // Always returns nil

	s2, err := decodeFromBuffer(buf)
	if err != nil {
		return err
	}
	s.hasAllSeq = s2.hasAllSeq
	s.subSeqSets = s2.subSeqSets
	s.permBag = s2.permBag

	return nil
}

func decodeFromBuffer(buf *bytes.Buffer) (*SeqSet, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == 128 {
		return ContainsAllSeqSet, nil
	}

	// If the SeqSet is a permutation.
	if b == 255 {
		permBagByte, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return &permutations[permBagByte], nil
	}

	s := new(SeqSet)
	for idx := 0; idx < len(s.subSeqSets); idx++ {
		isNil := b&(1<<uint(idx)) == 0
		if !isNil {
			seq, err := decodeFromBuffer(buf)
			if err != nil {
				return nil, err
			}
			s.subSeqSets[idx] = seq
		}
	}
	return s, nil
}

// Equals returns if two SeqSets are equal.
func (s *SeqSet) Equals(other *SeqSet) bool {
	b1, _ := s.GobEncode()
	b2, _ := other.GobEncode()
	return bytes.Equal(b1, b2)
}
