package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"tetris/combo4/bot"
	"time"
)

var (
	gobFile     = flag.String("gob_file", "mdp3.gob", "The path to a binary file of the MDP gob encoding or the path to write it to")
	previewLen  = flag.Int("preview_len", 3, "The number of pieces in preview")
	fromScratch = flag.Bool("from_scratch", false, "If set to true, does not read the MDP from file but creates a new one")
)

func main() {
	flag.Parse()

	start := time.Now()
	mdp := getMDP()
	fmt.Printf("Got initial MDP in %v\n", time.Since(start))

	mdp.Update(*gobFile)
	fmt.Printf("Completed in %v", time.Since(start))
}

func getMDP() *bot.MDP {
	// Create a new MDP.
	if *fromScratch {
		mdp, err := bot.NewMDP(*previewLen)
		if err != nil {
			fmt.Printf("NewMDP failed: %v\n", err)
			os.Exit(1)
		}
		return mdp
	}

	// Fetch the MDP from file.
	bytes, err := ioutil.ReadFile(*gobFile)
	if err != nil {
		fmt.Printf("failed to read file at %q: %v\n", *gobFile, err)
		fmt.Printf("Maybe try using --from_scratch")
		os.Exit(1)
	}
	mdp := &bot.MDP{}
	if err := mdp.GobDecode(bytes); err != nil {
		fmt.Printf("GobDecode failed: %v\n", err)
		os.Exit(1)
	}
	return mdp
}
