package bot

import (
	"math/rand"
	"testing"
	"tetris"
	"tetris/combo4"
)

func BenchmarkNextState(b *testing.B) {
	d := NewDecider()
	_, states := continuousNFAAndStates()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		randState := states[rand.Intn(len(states))]
		queue := tetris.RandPieces(7)
		var bag tetris.PieceSet

		d.NextState(randState, queue, bag)
	}
}

// Number of completed games 0.733% of 30 tries
// Average pieces played 311.900
func BenchmarkPlay400(b *testing.B) {
	d := NewDecider()
	start := combo4.State{Field: combo4.RightI}
	queues := make([][]tetris.Piece, b.N)
	for i := 0; i < b.N; i++ {
		queues[i] = tetris.RandPieces(408)
	}
	b.ResetTimer()

	var totalMoves, lost int
	for n := 0; n < b.N; n++ {
		fullQueue := queues[n]
		bag := tetris.NewPieceSet(fullQueue[:7]...)
		state := start
		for i := 0; i < len(fullQueue)-8; i++ {
			queue := fullQueue[i : i+7]

			var err error
			state, err = d.NextState(state, queue, bag)
			if err != nil {
				lost++
				break
			}
			totalMoves++

			if bag.Len() == 7 {
				bag = 0
			}
			bag = bag.Add(fullQueue[i+7])
		}
	}
	b.Logf("Number of completed games %.3f%% of %d tries", float64(b.N-lost)/float64(b.N), b.N)
	b.Logf("Average pieces played %.3f", float64(totalMoves)/float64(b.N))
}
