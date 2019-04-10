// Package main is used to compare different Scorers.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"tetris"
	"tetris/combo4"
	"tetris/combo4/bot"
	"text/tabwriter"
	"time"
)

var (
	numTrials     = flag.Int("num_trials", 1000, "the number of trials to test each scorer with")
	previewSize   = flag.Int("preview_size", 6, "the number of pieces you can see in the preview")
	deterministic = flag.Bool("deterministic", true, "whether the output is the same with each run")
)

// Which points to keep track of.
var checkpoints = [...]int{100, 500, 1000, 2000, 5000, 10000}

// The Scorers to test.
var scorersWithNames = [...]struct {
	scorer bot.Scorer
	name   string
}{
	{new(bot.NumStatesScorer), "# States"},
	{bot.NewNFAScorer(2), "Seq 2"},
	{bot.NewNFAScorer(4), "Seq 4"},
	{bot.NewNFAScorer(7), "Seq 7"},
}

func main() {
	flag.Parse()

	if !*deterministic {
		rand.Seed(time.Now().UnixNano())
	}

	// TODO(benjaminchang): Include non-continuous moves.
	nfa := combo4.NewNFA(combo4.AllContinuousMoves())

	var (
		totals [len(scorersWithNames)]int
		counts [len(scorersWithNames)][len(checkpoints)]int

		nfaTotal  int
		nfaCounts [len(checkpoints)]int
	)

	var deciders [len(scorersWithNames)]*bot.Decider
	for idx := range deciders {
		deciders[idx] = bot.NewDecider(scorersWithNames[idx].scorer)
	}

	piecesPerTrial := checkpoints[len(checkpoints)-1]

	for t := 0; t < *numTrials; t++ {
		if (t+1)%10 == 0 {
			fmt.Printf("Trial %d of %d\n", t+1, *numTrials)
		}
		queue := tetris.RandPieces(piecesPerTrial + *previewSize + 1)

		for dIdx, d := range deciders {
			dIdx, d := dIdx, d // Capture range variable.
			input := make(chan tetris.Piece, 1)

			output := d.StartGame(combo4.LeftI, queue[0], queue[1:*previewSize+1], input)
			if <-output == nil {
				break
			}

			consumed := 1
			for _, p := range queue[*previewSize+1:] {
				input <- p
				if <-output == nil {
					break
				}
				consumed++
				for cIdx, c := range checkpoints {
					if consumed == c {
						counts[dIdx][cIdx]++
					}
				}
			}
			totals[dIdx] += consumed
		}

		_, nfaCount := nfa.EndStates(combo4.NewStateSet(combo4.State{Field: combo4.LeftI}), queue)
		nfaTotal += nfaCount
		for cIdx, c := range checkpoints {
			if nfaCount > c {
				nfaCounts[cIdx]++
			}
		}
	}

	fmt.Printf("\n\nPreview Size = %d pieces\nTrials = %d\nMax sequence per trial = %d\n", *previewSize, *numTrials, piecesPerTrial)

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	title := "\tAvg"
	for _, c := range checkpoints {
		title += fmt.Sprintf("\tReach %d", c)
	}
	fmt.Fprintln(w, title)

	const fmtString = "\t%.1f%%"
	for idx := range deciders {
		row := scorersWithNames[idx].name
		row += fmt.Sprintf("\t%.1f", float64(totals[idx])/float64(*numTrials))
		for _, count := range counts[idx] {
			row += fmt.Sprintf(fmtString, float64(count*100)/float64(*numTrials))
		}
		fmt.Fprintln(w, row)
	}

	nfaRow := "Upper-bound"
	nfaRow += fmt.Sprintf("\t%.1f", float64(nfaTotal)/float64(*numTrials))
	for _, count := range nfaCounts {
		nfaRow += fmt.Sprintf(fmtString, float64(count*100)/float64(*numTrials))
	}
	fmt.Fprintln(w, nfaRow)

	w.Flush()
}
