package tetris

import (
	"bytes"
	"errors"
	"fmt"
)

// Seq represents a sequence of 7 or fewer pieces.
// Seq can be used as a map key. A seq cannot contain
// empty pieces.
type Seq struct {
	encoding uint32
	len      uint8
}

// NewSeq returns a Seq or an error if the length of the slice
// is over 7 or contains empty pieces.
func NewSeq(pieces []Piece) (Seq, error) {
	if len(pieces) > 7 {
		return Seq{}, errors.New("len(pieces) must be 7 or less")
	}
	var encoding uint32
	for idx, p := range pieces {
		if p == EmptyPiece {
			return Seq{}, errors.New("Seq cannot contain EmptyPiece")
		}
		encoding += uint32(p) << (4 * uint32(idx))
	}
	return Seq{encoding, uint8(len(pieces))}, nil
}

// MustSeq returns a new Seq and panics if the slice is over
// 7 in length.
func MustSeq(p []Piece) Seq {
	seq, err := NewSeq(p)
	if err != nil {
		panic(fmt.Sprintf("NewSeq failed: %v", err))
	}
	return seq
}

// ToSlice converts a Seq into a []Piece.
func (seq Seq) ToSlice() []Piece {
	slice := make([]Piece, seq.len)
	for idx := uint8(0); idx < seq.len; idx++ {
		shift := 4 * uint32(idx)
		slice[idx] = Piece((seq.encoding >> shift) & 15)
	}
	return slice
}

func (seq Seq) String() string {
	return fmt.Sprintf("%v", seq.ToSlice())
}

// Append returns a new Seq with the piece appended.
func (seq Seq) Append(p Piece) (Seq, error) {
	if seq.len >= 7 {
		return Seq{}, errors.New("Seq is already at max capacity")
	}
	return Seq{
		encoding: seq.encoding + uint32(p)<<(4*uint32(seq.len)),
		len:      seq.len + 1,
	}, nil
}

// SeqSet represents a set of sequences.
//
// The SeqSet is defined by prefixes. For example, given a SeqSet with the
// prefix [T O I], the SeqSet contains [T O I], [T O I T], [T O I Z Z], etc.
// This can be useful for storing Sequences that fail since you know that all
// sequences with that prefix will also fail.
//
// The empty struct is usable.
type SeqSet struct {
	hasAllSeq  bool       // Whether all sequences are contained.
	subSeqSets [7]*SeqSet // "map" from Piece to a SeqSet.
}

// hasAllSeqSet is used as a leaf node when possible.
var hasAllSeqSet = &SeqSet{hasAllSeq: true}

// AddPrefix adds the prefix to the SeqSet.
// AddPrefix panics if the prefix contains an EmptyPiece.
func (s *SeqSet) AddPrefix(prefix []Piece) {
	if len(prefix) == 0 {
		s.hasAllSeq = true
		// Zero out all the sub sequences sets which are now redundant.
		for i := 0; i < len(s.subSeqSets); i++ {
			s.subSeqSets[i] = nil
		}
		return
	}

	cur := s
	for _, piece := range prefix[:len(prefix)-1] {
		if cur.hasAllSeq {
			return // This prefix is already included.
		}
		next := cur.subSeqSets[piece-1]
		if next == nil {
			next = new(SeqSet)
			cur.subSeqSets[piece-1] = next
		}
		cur = next
	}

	lastPiece := prefix[len(prefix)-1]
	cur.subSeqSets[lastPiece-1] = hasAllSeqSet
}

// Contains return if the sequence is contained in the SeqSet.
// Contains panics if the sequence contains an EmptyPiece.
func (s *SeqSet) Contains(sequence []Piece) bool {
	if s == nil {
		return false
	}
	if s.hasAllSeq {
		return true
	}
	if len(sequence) == 0 {
		return false
	}
	sub := s.subSeqSets[sequence[0]-1]
	return sub.Contains(sequence[1:])
}

// Size returns the total number of sequences of a given length in the SeqSet.
func (s *SeqSet) Size(length int) int {
	if s == nil {
		return 0
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
func (s SeqSet) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	s.encodeToBuffer(buf)
	return buf.Bytes(), nil
}

func (s SeqSet) encodeToBuffer(buf *bytes.Buffer) {
	var b uint8
	if s.hasAllSeq {
		b |= 1 << 7
	}
	// Capture which indices are null.
	for idx, sub := range s.subSeqSets {
		if sub != nil {
			b |= 1 << uint(idx)
		}
	}
	buf.WriteByte(b) // Always returns nil

	// Write all the sub buffers.
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
	return nil
}

func decodeFromBuffer(buf *bytes.Buffer) (*SeqSet, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if hasAllSeq := b&(1<<7) != 0; hasAllSeq {
		return hasAllSeqSet, nil
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
	if s == nil || other == nil {
		return s == nil && other == nil
	}
	b1, _ := s.GobEncode()
	b2, _ := other.GobEncode()
	return bytes.Equal(b1, b2)
}
