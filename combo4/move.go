package combo4

import (
	"fmt"
	"tetris"
)

// Move represents a move that can be made to continue a 4 wide combo.
type Move struct {
	Start Field4x4
	End   Field4x4
	// The piece that is used to transition from start to end.
	Piece tetris.Piece
}

func (m Move) String() string {
	return fmt.Sprintf("{\nStart:\n%v\nEnd:\n%v\nPiece: %v\n}\n", m.Start, m.End, m.Piece)
}

// Cached list of actions for each move.
// Actions are specific to NullpoMino.
var actionsMap = make(map[Move][]tetris.Action)

// Actions returns the actions that must be done to perform the move. These
// are specific to NullpoMino.
func (m Move) Actions() []tetris.Action {
	if cached, ok := actionsMap[m]; ok {
		return cached
	}
	// TODO(benjaminchang): Add to the cache.
	return nil
}

type moveActions struct {
	Start Field4x4
	End   Field4x4
	Piece tetris.Piece

	// Actions that should be performed. Specific to NullpoMino
	Actions []tetris.Action
}

// AllContinuousMoves returns all moves that result in further play.
// See https://harddrop.com/wiki/Combo_Setups#4-Wide_with_3_Residua.
//
// AllContinousMoves also returns a set of actions that be done to
// execute the move. These actions apply to NullpoMino only.
func AllContinuousMoves() ([]Move, map[Move][]tetris.Action) {
	withoutReflect := make([]*moveActions, 0, 70)

	const X, o = true, false

	wallKickRight := []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right, tetris.SoftDrop, tetris.RotateCW}

	// Add moves excluding reflection.
	start := NewField4x4([][4]bool{
		{X, X, o, o},
		{X, o, o, o}})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, o, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.T,
			Actions: wallKickRight,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece:   tetris.L,
			Actions: wallKickRight,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.S,
			Actions: wallKickRight,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{X, o, o, X},
			}),
			Piece:   tetris.S,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, X, o},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, o},
				{X, o, o, o},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece:   tetris.O,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, o, X}})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{X, o, o, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, o, X},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, X, o, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, o, o, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece:   tetris.S,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.O,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, X, o, o}})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, X, o, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, o, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, X, o, o},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece:   tetris.O,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, X, o}})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, X, X},
			}),
			Piece: tetris.T,
			// T-spin bonus.
			Actions: []tetris.Action{tetris.Right, tetris.SoftDrop, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, X},
				{o, o, o, X},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{o, o, X, X},
			}),
			Piece:   tetris.S,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, o, X},
				{o, o, o, X},
			}),
			Piece:   tetris.I,
			Actions: []tetris.Action{tetris.RotateCW, tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, o, o},
		{X, o, o, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, o, X, o},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, o, o, X},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, o, o, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, o, X},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{o, o, X, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.RotateCW, tetris.RotateCCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, X},
				{o, o, X, o},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, X, X},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW},
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, o, o, X},
		{X, X, o, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.RotateCW, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW, tetris.SoftDrop, tetris.RotateCCW},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, o, o},
		{o, X, o, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, o, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, X, o},
			}),
			Piece:   tetris.Z,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.O,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, X, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, X, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, o, o},
		{X, X, o, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, o, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCCW, tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.O,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{o, X, X, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, X, o},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.J,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW, tetris.RotateCW},
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{o, X, o, X},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.T,
			Actions: []tetris.Action{tetris.Right, tetris.RotateCW, tetris.RotateCW},
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, o, X},
			}),
			Piece:   tetris.L,
			Actions: []tetris.Action{tetris.Right},
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, o, o},
		{X, o, o, X},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.S,
			Actions: []tetris.Action{tetris.RotateCW, tetris.SoftDrop, tetris.RotateCCW},
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, X, o},
		{X, o, o, o},
	})
	withoutReflect = append(withoutReflect, []*moveActions{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece:   tetris.L,
			Actions: wallKickRight,
		},
	}...)

	// Add an hard drop at the end of everything.
	for _, m := range withoutReflect {
		if l := len(m.Actions); l == 0 || m.Actions[l-1] != tetris.HardDrop {
			m.Actions = append(m.Actions, tetris.HardDrop)
		}
	}

	moves := make([]Move, 0, len(withoutReflect)*2)
	actions := make(map[Move][]tetris.Action, len(withoutReflect)*2)

	for _, m := range withoutReflect {
		move := Move{
			Start: m.Start,
			End:   m.End,
			Piece: m.Piece,
		}
		moves = append(moves, move)
		actions[move] = m.Actions
	}

	// Add the reflection of all the current moves.
	for _, unreflected := range withoutReflect {
		move := Move{
			Start: unreflected.Start.Mirror(),
			End:   unreflected.End.Mirror(),
			Piece: unreflected.Piece.Mirror(),
		}
		moves = append(moves, move)

		var mirrActions []tetris.Action
		// All pieces spawn on off center except (bias torwards the left)
		// except for I and O.
		switch move.Piece {
		case tetris.I, tetris.O:
			mirrActions = mirrorActions(unreflected.Actions)
		default:
			if unreflected.Actions[0] == tetris.Right {
				mirrActions = mirrorActions(unreflected.Actions[1:])
				break
			}
			mirrActions = make([]tetris.Action, len(unreflected.Actions)+1)
			mirrActions = append(mirrActions, tetris.Left)
			mirrActions = append(mirrActions, mirrorActions(unreflected.Actions)...)
		}
		actions[move] = mirrActions
	}

	return moves, actions
}

func mirrorActions(acts []tetris.Action) []tetris.Action {
	mirror := make([]tetris.Action, 0, len(acts))
	for _, a := range acts {
		mirror = append(mirror, a.Mirror())
	}
	return mirror
}
