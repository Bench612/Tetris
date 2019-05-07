// This packages generates a policy.MDP object and saves it to a file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"tetris/combo4/policy"
	"time"
)

var (
	gobFile     = flag.String("mdp_file", "mdp5.gob", "The path to a binary file of the MDP gob encoding or the path to write it to")
	previewLen  = flag.Int("preview_len", 5, "The number of pieces in preview")
	maxCombo    = flag.Int("max_combo", -1, "The maximum combo")
	fromScratch = flag.Bool("from_scratch", false, "If set to true, does not read the MDP from file but creates a new one")
)

func main() {
	flag.Parse()

	start := time.Now()
	mdp := getMDP()
	fmt.Printf("Got initial MDP in %v\n", time.Since(start))

	if err := mdp.Update(*gobFile); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}
	fmt.Printf("Completed in %v", time.Since(start))
}

func getMDP() *policy.MDP {
	// Create a new MDP.
	if *fromScratch {
		mdp, err := policy.NewMDP(*previewLen)
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
		fmt.Println("Maybe try using --from_scratch")
		os.Exit(1)
	}
	mdp := &policy.MDP{}
	if err := mdp.GobDecode(bytes); err != nil {
		fmt.Printf("GobDecode failed: %v\n", err)
		os.Exit(1)
	}
	return mdp
}
