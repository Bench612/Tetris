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

	// actionLimit is used to iterate through all actions.
	actionLimit
)

func (a Action) String() string {
	switch a {
	case NoAction:
		return "No_Action"
	case Hold:
		return "Swap_Hold"
	case Left:
		return "Left"
	case Right:
		return "Right"
	case RotateCW:
		return "Rotate_CW"
	case RotateCCW:
		return "Rotate_CCW"
	case SoftDrop:
		return "Soft_Drop"
	case HardDrop:
		return "Hard_Drop"
	}
	return "Unknown"
}

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
