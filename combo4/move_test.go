package combo4

import (
	"fmt"
	"testing"
	"tetris"

	"github.com/google/go-cmp/cmp"
)

func TestAllContinuousMoves(t *testing.T) {
	all := AllContinuousMoves()

	// Verify there are the right number of moves for each piece.
	pieceCount := make(map[tetris.Piece]int)
	for _, move := range all {
		pieceCount[move.Piece]++
	}
	wantPieceCount := map[tetris.Piece]int{
		tetris.T: 24,
		tetris.L: 27,
		tetris.J: 27,
		tetris.S: 11,
		tetris.Z: 11,
		tetris.O: 10,
		tetris.I: 30,
	}
	if diff := cmp.Diff(wantPieceCount, pieceCount); diff != "" {
		t.Errorf("piece count mismatch(-want +got):\n%s", diff)
	}

	// Verify that each move is valid.
	for _, move := range all {
		if err := checkMove(t, move); err != nil {
			t.Errorf("move %+v is invalid: %v", move, err)
		}
	}

	// Verify that nothing is repeated.
	moveCount := make(map[Move]int)
	for _, move := range all {
		moveCount[move]++
	}
	for move, count := range moveCount {
		if count != 1 {
			t.Errorf("move %+v occurs %d times, want 1", move, count)
		}
	}

}

// checkMove returns an error if the move is invalid.
func checkMove(t *testing.T, move Move) error {
	if got := move.Start.NumOccupied(); got != 3 {
		return fmt.Errorf("%d spaces occupied in the Start, want 3", got)
	}
	if got := move.End.NumOccupied(); got != 3 {
		return fmt.Errorf("%d spaces occupied in the End, want 3", got)
	}

	endArr := move.End.Array2D()
	fullRow := [4]bool{true, true, true, true}
	// Figure out possible end states before a row was cleared.
	preclearFields := []Field4x4{
		NewField4x4([][4]bool{fullRow, endArr[1], endArr[2], endArr[3]}),
		NewField4x4([][4]bool{endArr[1], fullRow, endArr[2], endArr[3]}),
		NewField4x4([][4]bool{endArr[1], endArr[2], fullRow, endArr[3]}),
		NewField4x4([][4]bool{endArr[1], endArr[2], endArr[3], fullRow}),
	}
	var validPiece tetris.Piece
	for _, preclear := range preclearFields {
		// Clear the start pieces. If this is the correct preclear field,
		// the remaining blocks should form a piece.
		pieceField := preclear &^ move.Start
		if pieceField.NumOccupied() != 4 {
			continue
		}
		pieceField, _, _ = toCanonicalPieceField(pieceField)
		switch p := canonicalPieceMap[pieceField]; p {
		case tetris.EmptyPiece:
		case move.Piece:
			return nil
		default:
			validPiece = p
		}
	}
	if validPiece != tetris.EmptyPiece {
		return fmt.Errorf("there is no transition from start -> end using %s but there is one using %s", move.Piece, validPiece)
	}
	return fmt.Errorf("there is no transition from start -> end using %s", move.Piece)
}

func toCanonicalPieceField(f Field4x4) (canonical Field4x4, rowShift int, colShift int) {
	arr := f.Array2D()
	maxRow := -1
	minCol := 4
	for rowIdx, row := range arr {
		for colIdx, isSet := range row {
			if !isSet {
				continue
			}
			if rowIdx > maxRow {
				maxRow = rowIdx
			}
			if colIdx < minCol {
				minCol = colIdx
			}
		}
	}
	var shiftedArr [4][4]bool
	rowShift = 3 - maxRow
	colShift = -minCol
	for r := 0; r <= maxRow; r++ {
		for c := 3; c >= minCol; c-- {
			shiftedArr[r+rowShift][c+colShift] = arr[r][c]
		}
	}
	return NewField4x4(shiftedArr[:]), rowShift, colShift
}

// The canonicalPieceMap is a map from Field4x4 of all rotations of pieces to the piece.
// The canonicalPieceMap places pieces in the bottom left.
var canonicalPieceMap = map[Field4x4]tetris.Piece{
	// T
	NewField4x4([][4]bool{
		{false, true, false, false},
		{true, true, true, false},
	}): tetris.T,
	NewField4x4([][4]bool{
		{true, false, false},
		{true, true, false},
		{true, false, false},
	}): tetris.T,
	NewField4x4([][4]bool{
		{true, true, true, false},
		{false, true, false, false},
	}): tetris.T,
	NewField4x4([][4]bool{
		{false, true, false, false},
		{true, true, false, false},
		{false, true, false, false},
	}): tetris.T,
	//L
	NewField4x4([][4]bool{
		{true, false, false, false},
		{true, false, false, false},
		{true, true, false, false},
	}): tetris.L,
	NewField4x4([][4]bool{
		{true, true, true, false},
		{true, false, false, false},
	}): tetris.L,
	NewField4x4([][4]bool{
		{true, true, false, false},
		{false, true, false, false},
		{false, true, false, false},
	}): tetris.L,
	NewField4x4([][4]bool{
		{false, false, true, false},
		{true, true, true, false},
	}): tetris.L,
	// J
	NewField4x4([][4]bool{
		{true, true, false, false},
		{true, false, false, false},
		{true, false, false, false},
	}): tetris.J,
	NewField4x4([][4]bool{
		{true, true, true, false},
		{false, false, true, false},
	}): tetris.J,
	NewField4x4([][4]bool{
		{false, true, false, false},
		{false, true, false, false},
		{true, true, false, false},
	}): tetris.J,
	NewField4x4([][4]bool{
		{true, false, false, false},
		{true, true, true, false},
	}): tetris.J,
	// S
	NewField4x4([][4]bool{
		{false, true, true, false},
		{true, true, false, false},
	}): tetris.S,
	NewField4x4([][4]bool{
		{true, false, false, false},
		{true, true, false, false},
		{false, true, false, false},
	}): tetris.S,
	// Z
	NewField4x4([][4]bool{
		{true, true, false, false},
		{false, true, true, false},
	}): tetris.Z,
	NewField4x4([][4]bool{
		{false, true, false, false},
		{true, true, false, false},
		{true, false, false, false},
	}): tetris.Z,
	// O
	NewField4x4([][4]bool{
		{true, true, false, false},
		{true, true, false, false},
	}): tetris.O,
	// I
	NewField4x4([][4]bool{{true, true, true, true}}):       tetris.I,
	NewField4x4([][4]bool{{true}, {true}, {true}, {true}}): tetris.I,
}
