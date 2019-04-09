package tetris

import (
	"fmt"
	"sort"
)

func init() {
	// Generate the permutations.
	for _, bag := range AllPieceSets() {
		bagIdx := bag
		permutations[bagIdx].isPermutation = true

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
// SeqSets cannot contain EmptyPieces and most functions will panic if an
// EmptyPiece is supplied.
//
// SeqSets are immutable. The nil pointer is usable.
type SeqSet struct {
	// "map" from (Piece-1) to a SeqSet.
	subSeqSets [7]*SeqSet
	// Whether the SeqSet is from the global permutations var.
	isPermutation bool
}

// ContainsAllSeqSet is a special SeqSet that contains all sequences.
var ContainsAllSeqSet = &SeqSet{}

// permutations is a collection of special SeqSets from PieceSet->SeqSet.
// Each SeqSet contains all possible permutations from the given bag state.
// These SeqSets are special because only these SeqSets, SeqSets containing these
// have cycles.
var (
	permutations [255]SeqSet
)

// Permutations returns a SeqSet that contains all sequences starting from the
// given bag state assuming a 7 bag randomizer.
func Permutations(bagUsed PieceSet) *SeqSet {
	return &permutations[bagUsed]
}

// NewSeqSet contructs a new SeqSet from a list of prefixes.
func NewSeqSet(prefixes ...[]Piece) *SeqSet {
	if len(prefixes) == 0 {
		return nil
	}
	sorted := make([][]Piece, len(prefixes))
	copy(sorted, prefixes)
	sort.Slice(sorted, func(i, j int) bool { return len(sorted[i]) < len(sorted[j]) })

	s := new(SeqSet)
	for _, prefix := range sorted {
		if len(prefix) == 0 {
			return ContainsAllSeqSet
		}
		s.addPrefix(prefix)
	}
	return s
}

// addPrefix adds the prefix to the SeqSet. Assumes prefix is at least length 1.
// This should only be called in the constructor to keep SeqSets immutable.
func (s *SeqSet) addPrefix(prefix []Piece) {
	if s == ContainsAllSeqSet {
		return
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
	next.addPrefix(prefix[1:])
}

// PrependedSeqSets can be used to construct a SeqSet from other SeqSets.
// For example, given a set [[I,O,J], [I,J]], you can create a set that pre
// pre-pends S to each sequence to get [[S,I,O,J], [S,I,J]].
//
// The prefixToSet arg is a "map" from Piece to SeqSet.
func PrependedSeqSets(prefixToSet [8]*SeqSet) *SeqSet {
	s := new(SeqSet)
	copy(s.subSeqSets[:], prefixToSet[1:])
	return s
}

// Contains returns if the sequence is contained in the SeqSet.
// Contains panics if the sequence contains an EmptyPiece.
func (s *SeqSet) Contains(sequence []Piece) bool {
	if s == nil {
		return false
	}
	if s == ContainsAllSeqSet {
		return true
	}
	if len(sequence) == 0 {
		// Permutations contain all sequences that dont lead to nil.
		return s.isPermutation
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
	if s == nil || s.isPermutation {
		return nil
	}
	if s == ContainsAllSeqSet {
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
	if s == ContainsAllSeqSet {
		return "{prefixes=all}"
	}
	return fmt.Sprintf("{prefixes=%v}", s.Prefixes())
}

// Union returns the union of this SeqSet and another.
func (s *SeqSet) Union(other *SeqSet) *SeqSet {
	if s == nil {
		return other
	}
	if other == nil {
		return s
	}
	if s == ContainsAllSeqSet || other == ContainsAllSeqSet {
		return ContainsAllSeqSet
	}
	union := &SeqSet{}
	for i := range union.subSeqSets {
		union.subSeqSets[i] = s.subSeqSets[i].Union(other.subSeqSets[i])
	}
	return union
}

// Intersection returns the intersection of this SeqSet and another.
func (s *SeqSet) Intersection(other *SeqSet) *SeqSet {
	if s == nil || other == nil {
		return nil
	}
	if s == ContainsAllSeqSet {
		return other
	}
	if other == ContainsAllSeqSet {
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
	if s.isPermutation {
		if length == 0 {
			return 1
		}
		// Calculate the number of sequences by the choices at each step
		// assuming a 7 bag randomizer.
		choices := 0
		for _, sub := range s.subSeqSets {
			if sub != nil {
				choices++
			}
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
	if s == ContainsAllSeqSet {
		// 7^length
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

// Equals returns true if two SeqSets are equivalent.
func (s *SeqSet) Equals(other *SeqSet) bool {
	if s == nil || other == nil {
		return s == nil && other == nil
	}
	if s == other {
		return true
	}
	for idx := range s.subSeqSets {
		if (s.subSeqSets[idx] == nil && other.subSeqSets[idx] != nil) ||
			(s.subSeqSets[idx] != nil && other.subSeqSets[idx] == nil) {
			return false
		}
	}
	for idx := range s.subSeqSets {
		if !s.subSeqSets[idx].Equals(other.subSeqSets[idx]) {
			return false
		}
	}
	return true
}
