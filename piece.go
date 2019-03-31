package tetris

import (
	"errors"
	"fmt"
	"math/bits"
	"math/rand"
)

// Piece represents a tetrimino or empty piece.
type Piece uint8

// Possible pieces.
const (
	EmptyPiece Piece = iota
	T
	L
	J
	S
	Z
	O
	I
)

// NonemptyPieces is an ordered array of non-empty pieces.
var NonemptyPieces = [7]Piece{T, L, J, S, Z, O, I}

func (p Piece) String() string {
	switch p {
	case EmptyPiece:
		return "Ɛ"
	case T:
		return "T"
	case L:
		return "L"
	case J:
		return "J"
	case S:
		return "S"
	case Z:
		return "Z"
	case O:
		return "O"
	case I:
		return "I"
	}
	panic("Unknown piece")
}

// GameString returns a string depiction of what the piece looks like.
func (p Piece) GameString() string {
	switch p {
	case EmptyPiece:
		return ""
	case T:
		return "_□_\n□□□"
	case L:
		return "__□\n□□□"
	case J:
		return "□__\n□□□"
	case S:
		return "_□□\n□□_"
	case Z:
		return "□□_\n_□□"
	case O:
		return "□□\n□□"
	case I:
		return "□□□□"
	}
	panic("Unknown piece")
}

// PieceSet returns a PieceSet containing only this Piece.
func (p Piece) PieceSet() PieceSet {
	return 1 << p
}

// Mirror returns the mirrored version of a piece.
func Mirror(p Piece) Piece {
	switch p {
	case L:
		return J
	case J:
		return L
	case S:
		return Z
	case Z:
		return S
	}
	return p
}

// RandPieces turns a slice of random pieces using a 7 bag randomizer.
func RandPieces(length int) []Piece {
	pieces := make([]Piece, 0, length+6)
	for len(pieces) < length {
		for _, i := range rand.Perm(7) {
			pieces = append(pieces, Piece(i+1))
		}
	}
	return pieces[:length]
}

// PieceSet represents a set of pieces. Duplicates and EmptyPieces are not recorded.
// The empty value is represents no pieces.
type PieceSet uint8

// NewPieceSet creates a new PieceSet from the specified Pieces.
func NewPieceSet(pieces ...Piece) PieceSet {
	var ps PieceSet
	for _, p := range pieces {
		ps |= p.PieceSet()
	}
	// Zero out the EmptyPiece.
	ps &^= EmptyPiece.PieceSet()
	return ps
}

// Union returns the union of two PieceSets.
func (ps PieceSet) Union(other PieceSet) PieceSet {
	return ps | other
}

// Add returns a PieceSet with a certain Piece added.
func (ps PieceSet) Add(p Piece) PieceSet {
	return ps | p.PieceSet()
}

// Contains returns whether the PieceSet contains the piece.
func (ps PieceSet) Contains(p Piece) bool {
	return ps&p.PieceSet() != 0
}

// Len returns the number of items in the PieceSet.
func (ps PieceSet) Len() int {
	return bits.OnesCount8(uint8(ps))
}

// ToSlice returns a slice of all Pieces represented by this set.
func (ps PieceSet) ToSlice() []Piece {
	if ps.Len() == 0 {
		return nil
	}
	slice := make([]Piece, 0, ps.Len())
	for _, piece := range NonemptyPieces {
		if ps.Contains(piece) {
			slice = append(slice, piece)
		}
	}
	return slice
}

func (ps PieceSet) String() string {
	return fmt.Sprint(ps.ToSlice())
}

// Inverted returns a PieceSet that contains all Pieces *not* contained in this
// PieceSet.
func (ps PieceSet) Inverted() PieceSet {
	// Invert and zero out the EmptyPiece.
	return (ps ^ 255) &^ (1 << EmptyPiece)
}

// PieceSeq represents a sequence of 7 or fewer pieces.
// PieceSeq can be used as a map key.
type PieceSeq struct {
	encoding uint32
	len      uint8
}

// NewPieceSeq returns a PieceSeq or an error if the length of the slice
// is over 7.
func NewPieceSeq(pieces []Piece) (PieceSeq, error) {
	if len(pieces) > 7 {
		return PieceSeq{}, errors.New("len(pieces) must be 7 or less")
	}
	var encoding uint32
	for idx, p := range pieces {
		encoding += uint32(p) << (4 * uint32(idx))
	}
	return PieceSeq{encoding, uint8(len(pieces))}, nil
}

// MustPieceSeq returns a new PieceSeq and panics if the slice is over
// 7 in length.
func MustPieceSeq(p []Piece) PieceSeq {
	seq, err := NewPieceSeq(p)
	if err != nil {
		panic(fmt.Sprintf("NewPieceSeq failed: %v", err))
	}
	return seq
}

// ToSlice converts a PieceSeq into a []Piece.
func (seq PieceSeq) ToSlice() []Piece {
	slice := make([]Piece, seq.len)
	for idx := uint8(0); idx < seq.len; idx++ {
		shift := 4 * uint32(idx)
		slice[idx] = Piece((seq.encoding >> shift) & 15)
	}
	return slice
}

// Append returns a new PieceSeq with the piece appended.
func (seq PieceSeq) Append(p Piece) (PieceSeq, error) {
	if seq.len >= 7 {
		return PieceSeq{}, errors.New("PieceSeq is already at max capacity")
	}
	return PieceSeq{
		encoding: seq.encoding + uint32(p)<<(4*uint32(seq.len)),
		len:      seq.len + 1,
	}, nil
}
