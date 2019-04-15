package tetris

import (
	"errors"
	"fmt"
)

// Seq represents a sequence of 8 or fewer pieces.
// Seq can be used as a map key.
type Seq uint32

// NewSeq returns a Seq or an error if the length of the slice
// is over 8.
func NewSeq(pieces []Piece) (Seq, error) {
	if len(pieces) > 8 {
		return 0, errors.New("len(pieces) must be 8 or less")
	}
	var seq uint32
	for idx, p := range pieces {
		if p == EmptyPiece {
			return 0, errors.New("Seq cannot contain EmptyPiece")
		}
		seq += uint32(p) << (uint32(idx) << 2)
	}
	return Seq(seq), nil
}

// MustSeq returns a new Seq and panics if the slice is over
// 8 in length.
func MustSeq(p []Piece) Seq {
	seq, err := NewSeq(p)
	if err != nil {
		panic(fmt.Sprintf("NewSeq failed: %v", err))
	}
	return seq
}

// Slice converts a Seq into a []Piece.
func (seq Seq) Slice() []Piece {
	if seq == 0 {
		return nil
	}
	slice := make([]Piece, 0, 7)
	for idx := 0; ; idx++ {
		p := seq.AtIndex(idx)
		if p == EmptyPiece {
			break
		}
		slice = append(slice, p)
	}
	return slice
}

// AtIndex returns what piece is at the index of the Sequence or EmptyPiece.
func (seq Seq) AtIndex(idx int) Piece {
	shift := uint(idx) << 2
	return Piece((seq >> shift) & 15)
}

// SetIndex returns a Seq with a the piece set at the specified index.
func (seq Seq) SetIndex(idx int, p Piece) Seq {
	if idx < 0 || (idx > 0 && seq.AtIndex(idx-1) == EmptyPiece) || idx >= 8 {
		panic("index out of bounds")
	}
	shift := uint(idx) << 2
	return seq&^(15<<shift) | Seq(p)<<shift
}

// RemoveFirst returns a new Seq that removes the first element from the Seq.
func (seq Seq) RemoveFirst() Seq {
	return seq >> 4
}

func (seq Seq) String() string {
	return fmt.Sprintf("%v", seq.Slice())
}
