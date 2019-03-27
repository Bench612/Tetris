package tetris

import (
	"fmt"
	"math/bits"
	"math/rand"
)

func init() {
	precomputeSlices()
}

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

// PieceSet represents a set of pieces.
// Duplicates and EmptyPieces are not recorded.
type PieceSet uint8

// NewPieceSet creates a new PieceSet from the specified Pieces.
func NewPieceSet(pieces ...Piece) PieceSet {
	var ps PieceSet
	for _, p := range pieces {
		ps |= 1 << p
	}
	// Zero out the EmptyPiece.
	ps &^= 1 << EmptyPiece
	return ps
}

// Union returns the union of two PieceSets.
func (ps PieceSet) Union(other PieceSet) PieceSet {
	return ps | other
}

// Add returns a PieceSet with a certain Piece added.
func (ps PieceSet) Add(p Piece) PieceSet {
	return ps | (1 << p)
}

// Contains returns whether the PieceSet contains the piece.
func (ps PieceSet) Contains(p Piece) bool {
	return ps&(1<<p) != 0
}

// Len returns the number of items in the PieceSet.
func (ps PieceSet) Len() int {
	return bits.OnesCount8(uint8(ps))
}

// Precompute the slices for each PieceSet.
var toSlices [256][]Piece

func precomputeSlices() {
	for i := 0; i <= 255; i++ {
		ps := PieceSet(i)
		toSlices[i] = ps.toSlice()
	}
}

// toSlice is the non-precomputed version of ToSlice.
func (ps PieceSet) toSlice() []Piece {
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

// ToSlice returns a slice of all Pieces represented by this set.
// This slice should not be modified.
func (ps PieceSet) ToSlice() []Piece {
	return toSlices[int(ps)]
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
