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

// AllContinuousMoves returns all moves that result in further play.
// See https://harddrop.com/wiki/Combo_Setups#4-Wide_with_3_Residua.
func AllContinuousMoves() []Move {
	all := make([]Move, 0, 140)

	const X, o = true, false

	// Add moves excluding reflection.
	start := NewField4x4([][4]bool{
		{X, X, o, o},
		{X, o, o, o}})
	all = append(all, []Move{
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
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.S,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{X, o, o, X},
			}),
			Piece: tetris.S,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, X, o},
			}),
			Piece: tetris.Z,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, o},
				{X, o, o, o},
			}),
			Piece: tetris.Z,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece: tetris.O,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, o, X}})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{X, o, o, X},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, o, X},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, o, o, X},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece: tetris.S,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.O,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, X, o, o}})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, X, o, o},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, o, X},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, o, X},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.Z,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece: tetris.O,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, X, o}})
	all = append(all, []Move{
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
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, X},
				{o, o, o, X},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{o, o, X, X},
			}),
			Piece: tetris.S,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, o, X},
				{o, o, o, X},
			}),
			Piece: tetris.I,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, o, o},
		{X, o, o, o},
	})
	all = append(all, []Move{
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
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, o, o, X},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, o, o, X},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, o, o},
				{X, X, o, o},
			}),
			Piece: tetris.J,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, o, X},
	})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, o},
				{o, o, X, X},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, X, X},
				{o, o, X, o},
			}),
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, o, X, X},
			}),
			Piece: tetris.Z,
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, o, o, X},
		{X, X, o, o},
	})
	all = append(all, []Move{
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
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.Z,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, X, o, o},
		{o, X, o, o},
	})
	all = append(all, []Move{
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
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, X, o},
			}),
			Piece: tetris.Z,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.O,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{X, o, X, o},
	})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, X, X, o},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{X, o, X, o},
			}),
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{X, o, X, X},
			}),
			Piece: tetris.J,
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, o, o},
		{X, X, o, o},
	})
	all = append(all, []Move{
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
			Piece: tetris.J,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.O,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{o, X, X, o},
	})
	all = append(all, []Move{
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
			Piece: tetris.L,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.J,
		},
	}...)
	start = NewField4x4([][4]bool{
		{X, o, o, o},
		{o, X, o, X},
	})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.T,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, o, o, X},
				{o, X, o, X},
			}),
			Piece: tetris.L,
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, o, o},
		{X, o, o, X},
	})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.S,
		},
	}...)
	start = NewField4x4([][4]bool{
		{o, X, X, o},
		{X, o, o, o},
	})
	all = append(all, []Move{
		{
			Start: start,
			End:   start,
			Piece: tetris.I,
		}, {
			Start: start,
			End: NewField4x4([][4]bool{
				{o, X, X, X},
			}),
			Piece: tetris.L,
		},
	}...)

	// Add the reflection of all the current moves.
	withoutReflect := len(all)
	for i := 0; i < withoutReflect; i++ {
		move := all[i]
		all = append(all, Move{
			Start: MirrorField4x4(move.Start),
			End:   MirrorField4x4(move.End),
			Piece: tetris.Mirror(move.Piece),
		})
	}

	return all
}
