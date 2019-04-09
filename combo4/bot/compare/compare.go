package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"tetris"
	"tetris/combo4"
	"tetris/combo4/bot"
	"text/tabwriter"
)

var numTrials = flag.Int("num_trials", 1000, "the number of trials to test each scorer with")
var nextSize = flag.Int("next_size", 6, "the number of pieces you can see in the next")

const piecesPerTrial = 2000

func main() {
	flag.Parse()

	fmt.Printf("num_trials=%d\n", *numTrials)

	// TODO(benjaminchang): Include non-continuous moves.
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())

	scorers := []bot.Scorer{
		new(bot.ZeroScorer),
		bot.NewNFAScorer(4),
		bot.NewNFAScorer(5),
		bot.NewNFAScorer(6),
		bot.NewNFAScorer(7),
		bot.NewNFAScorer(8),
	}
	names := []string{
		"NumStates",
		"Seq 4",
		"Seq 5",
		"Seq 6",
		"Seq 7",
		"Seq 8",
	}
	deciders := make([]*bot.Decider, len(scorers))
	for idx, s := range scorers {
		deciders[idx] = bot.NewDecider(s)
	}
	// The number of trials that make it past certain checkpoints.
	var (
		count20   = make([]int, len(scorers))
		count200  = make([]int, len(scorers))
		count2000 = make([]int, len(scorers))
	)
	// How far the NFA can go. That is, with optimal play and seeing all coming pieces.
	var (
		nfa20   int
		nfa200  int
		nfa2000 int
	)

	for t := 0; t < *numTrials; t++ {
		queue := tetris.RandPieces(piecesPerTrial + 8)

		var wg sync.WaitGroup
		wg.Add(len(deciders))
		for dIdx, d := range deciders {
			dIdx, d := dIdx, d // Capture range variables.
			go func() {
				defer wg.Done()

				input := make(chan tetris.Piece, 10)
				output := d.StartGame(combo4.LeftI, queue[0], queue[1:*nextSize+1], input)

				consumed := 0
				for _, p := range queue[*nextSize+1:] {
				playPieceLoop:
					for {
						select {
						case input <- p:
							break playPieceLoop
						case o := <-output:
							if o == nil {
								return
							}
							consumed++
							switch consumed {
							case 20:
								count20[dIdx]++
							case 200:
								count200[dIdx]++
							case 2000:
								count2000[dIdx]++
							}
						}
					}
				}
			}()
		}

		_, nfaCount := nfa.EndStates(combo4.NewStateSet(combo4.State{Field: combo4.LeftI}), queue)
		if nfaCount >= 20 {
			nfa20++
		}
		if nfaCount >= 200 {
			nfa200++
		}
		if nfaCount >= 2000 {
			nfa2000++
		}
		wg.Wait()
	}

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	fmt.Fprint(w, "\tReach 20\tReach 200\tReach 2000\n")

	const fmtStr = "%s\t%.2f%%\t%.2f%%\t%.2f%%\n"
	for idx := range deciders {
		fmt.Fprintf(w, fmtStr, names[idx],
			float64(count20[idx])*100/float64(*numTrials),
			float64(count200[idx])*100/float64(*numTrials),
			float64(count200[idx])*100/float64(*numTrials))
	}
	fmt.Fprintf(w, fmtStr, "Upper-bound",
		float64(nfa20)*100/float64(*numTrials),
		float64(nfa20)*100/float64(*numTrials),
		float64(nfa2000)*100/float64(*numTrials))
	w.Flush()
}
