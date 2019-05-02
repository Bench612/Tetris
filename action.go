package tetris

// Action represents something the user can do by pressing a key.
type Action uint8

// All possible actions.
const (
	NoAction Action = iota
	Hold
	Left
	Right
	RotateCW
	RotateCCW
	SoftDrop
	HardDrop
)

// Mirror returns the equivalent action if the field is reflected across the y
// axis.
func (a Action) Mirror() Action {
	switch a {
	case Left:
		return Right
	case Right:
		return Left
	case RotateCW:
		return RotateCCW
	case RotateCCW:
		return RotateCW
	}
	return a
}
