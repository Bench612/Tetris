package combo4

import "math/bits"

// Field4x4 represents the state of a 4x4 group of squares.
type Field4x4 uint16

// Common Field4x4s used to start a 4 wide combo.
const (
	LeftI  = 28672 // NewField4x4([][4]bool{{true, true, true, false}})
	RightI = 57344 // NewField4x4([][4]bool{{false, true, true, true}})
	LeftZ  = 12544 // NewField4x4([][4]bool{{true, false, false, false},{true, true, false, false}})
)

// NewField4x4 creates a new Field4x4. True represents an occupied space.
// If more than 4 rows are provided then only the bottom four rows will be
// considered. If fewer than 4 rows are provided, they will be placed at the
// bottom.
func NewField4x4(field [][4]bool) Field4x4 {
	var f4x4 uint16

	startI := len(field) - 1
	for i := startI; i >= 0 && i >= len(field)-4; i-- {
		rowVals := field[i]
		row := 3 - (startI - i)
		for col, isSet := range rowVals {
			if isSet {
				f4x4 |= 1 << uint((row*4)+col)
			}
		}
	}
	return Field4x4(f4x4)
}

// String returns a string representation of a field.
func (f Field4x4) String() string {
	runes := make([]rune, 0, 20)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if f.isSet(r, c) {
				runes = append(runes, '□')
			} else {
				runes = append(runes, '_')
			}
		}
		runes = append(runes, '\n')
	}
	return string(runes)
}

// Array2D returns a 2D boolean array represenation of the field.
func (f Field4x4) Array2D() [4][4]bool {
	var s [4][4]bool
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			s[r][c] = f.isSet(r, c)
		}
	}
	return s
}

// NumOccupied returns the number of squares that are taken.
func (f Field4x4) NumOccupied() int {
	return bits.OnesCount16(uint16(f))
}

// isSet returns if the specified row and column is occupied.
// isSet returns false for values out of bounds.
func (f Field4x4) isSet(row, col int) bool {
	if row >= 4 || col >= 4 || row < 0 || col < 0 {
		return false
	}
	var mask uint = 1 << uint((row*4)+col)
	return uint(f)&mask != 0
}

// MirrorField4x4 reflects a Field4x4 across the y axis through the middle.
func MirrorField4x4(f Field4x4) Field4x4 {
	array := f.Array2D()
	mirrored := make([][4]bool, 0, 4)
	for _, row := range array {
		mirrored = append(mirrored, [4]bool{row[3], row[2], row[1], row[0]})
	}
	return NewField4x4(mirrored)
}
