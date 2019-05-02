// Package main is used to compare different Scorers.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"tetris"
	"tetris/combo4"
	"tetris/combo4/bot"
	"text/tabwriter"
	"time"
)

var (
	numTrials     = flag.Int("num_trials", 200, "the number of trials to test each scorer with")
	previewSize   = flag.Int("preview_size", 6, "the number of pieces you can see in the preview")
	deterministic = flag.Bool("deterministic", true, "whether the output is the same with each run")
)

// Which points to keep track of.
var checkpoints = [...]int{100, 500, 1000, 2000, 5000, 10000, 20000, 30000}

var nfa = combo4.NewNFA(combo4.AllContinuousMoves())

// The Policies to test.
var policiesWithNames = [...]struct {
	name   string
	policy bot.Policy
}{
	{"Seq 3", bot.PolicyFromScorer(nfa, bot.NewNFAScorer(nfa, 3))},
	{"Seq 6", bot.PolicyFromScorer(nfa, bot.NewNFAScorer(nfa, 6))},
	{"MDP 6", newMDPPolicy("policy6.gob")},
}

func newMDPPolicy(path string) bot.Policy {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("ioutil.ReadFile(%q): %v\n", path, err)
		os.Exit(1)
	}
	mdpPol := &bot.MDPPolicy{}
	if err := mdpPol.GobDecode(bytes); err != nil {
		fmt.Printf("GobDecode failed: %v\n", err)
		os.Exit(1)
	}
	return mdpPol
}

/* Sample Output

Preview Size = 6 pieces
Trials = 200
Max sequence per trial = 30000
              Avg       Reach 100   Reach 500   Reach 1000   Reach 2000   Reach 5000   Reach 10000   Reach 20000   Reach 30000
Seq 3         587.2     67.0%       43.0%       21.5%        5.5%         0.0%         0.0%          0.0%          0.0%
Seq 6         1102.3    70.5%       56.5%       41.0%        18.0%        2.0%         0.0%          0.0%          0.0%
MDP 6         2420.9    73.5%       68.0%       57.0%        37.0%        15.0%        3.5%          0.5%          0.0%
Upper-bound   22717.4   77.0%       77.0%       77.0%        77.0%        77.0%        76.0%         75.0%         75.0%

*/
func main() {
	flag.Parse()

	if !*deterministic {
		rand.Seed(time.Now().UnixNano())
	}

	var (
		totals [len(policiesWithNames)]int
		counts [len(policiesWithNames)][len(checkpoints)]int

		nfaTotal  int
		nfaCounts [len(checkpoints)]int
	)

	piecesPerTrial := checkpoints[len(checkpoints)-1]

	// Add the totals and counts for each decider.
	type queueItem struct {
		dIdx     int
		consumed int
	}
	policiesCh := make(chan queueItem, 30)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < *numTrials*len(policiesWithNames); i++ {
			qItem := <-policiesCh
			for cIdx, c := range checkpoints {
				if qItem.consumed >= c {
					counts[qItem.dIdx][cIdx]++
				}
			}
			totals[qItem.dIdx] += qItem.consumed
		}
		wg.Done()
	}()

	// Add the totals and counts for the NFA
	nfaCh := make(chan int, 10)
	wg.Add(1)
	go func() {
		for i := 0; i < *numTrials; i++ {
			count := <-nfaCh
			nfaTotal += count
			for cIdx, c := range checkpoints {
				if count > c {
					nfaCounts[cIdx]++
				}
			}
		}
		wg.Done()
	}()

	maxConcurrency := make(chan bool, 32)
	for t := 0; t < *numTrials; t++ {
		if (t+1)%10 == 0 {
			fmt.Printf("Trial %d of %d\n", t+1, *numTrials)
		}
		queue := tetris.RandPieces(piecesPerTrial + *previewSize + 1)

		for dIdx, d := range policiesWithNames {
			dIdx, d := dIdx, d // Capture range variable.
			maxConcurrency <- true
			go func() {
				defer func() { <-maxConcurrency }()

				input := make(chan tetris.Piece, 1)

				output := bot.StartGame(d.policy, combo4.LeftI, queue[0], queue[1:*previewSize+1], input)
				var consumed int
				if <-output != nil {
					consumed++
					for _, p := range queue[*previewSize+1:] {
						input <- p
						if <-output == nil {
							break
						}
						consumed++
					}
				}
				policiesCh <- queueItem{dIdx: dIdx, consumed: consumed}
			}()
		}

		go func() {
			_, count := nfa.EndStates(combo4.NewStateSet(combo4.State{Field: combo4.LeftI}), queue)
			nfaCh <- count
		}()
	}

	// Wait for all trials to be computed.
	wg.Wait()

	fmt.Printf("\n\nPreview Size = %d pieces\nTrials = %d\nMax sequence per trial = %d\n", *previewSize, *numTrials, piecesPerTrial)

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	title := "\tAvg"
	for _, c := range checkpoints {
		title += fmt.Sprintf("\tReach %d", c)
	}
	fmt.Fprintln(w, title)

	const fmtString = "\t%.1f%%"
	for idx, d := range policiesWithNames {
		row := d.name
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
