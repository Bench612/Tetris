package bot

import (
	"tetris/combo4"
)

// Decider picks the next move
type Decider struct {
	nfa    *combo4.NFA
	scorer *Scorer
}
