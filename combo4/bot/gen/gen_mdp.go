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
	gobFile     = flag.String("gob_file", "mdp4.gob", "The path to a binary file of the MDP gob encoding or the path to write it to")
	previewLen  = flag.Int("preview_len", 4, "The number of pieces in preview")
	fromScratch = flag.Bool("from_scratch", false, "If set to true, does not read the MDP from file but creates a new one")
)

func main() {
	flag.Parse()

	var mdp *bot.MDP
	if *fromScratch {
		// Create a new MDP.
		var err error
		start := time.Now()
		mdp, err = bot.NewMDP(*previewLen)
		if err != nil {
			fmt.Printf("NewMDP failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created MDP in %v\n", time.Since(start))
	} else {
		// Fetch the current MDP.
		start := time.Now()
		bytes, err := ioutil.ReadFile(*gobFile)
		if err != nil {
			fmt.Printf("ioutil.ReadFile: %v\n", err)
			fmt.Printf("Maybe try using --from_scratch")
			os.Exit(1)
		}
		mdp = &bot.MDP{}
		if err := mdp.GobDecode(bytes); err != nil {
			fmt.Printf("GobDecode failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Fetched current MDP in %v\n", time.Since(start))
	}

	for cutoff := 5; ; cutoff *= 2 {
		if iterateUntilEquilibrium(mdp, cutoff) == 1 {
			iterateUntilEquilibrium(mdp, -1)
			break
		}
	}

	if err := save(mdp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func iterateUntilEquilibrium(mdp *bot.MDP, cutoff int) int {
	iter := 0
	for ; ; iter++ {
		// Now update the MDP iteratively.
		start := time.Now()
		var changed bool
		valueIters := 1
		for ; ; valueIters++ {
			updated := mdp.UpdateValues(cutoff)
			if updated == 0 {
				break
			}
			fmt.Printf("Updated %d values (#%d)\n", updated, valueIters)
			changed = true
		}
		fmt.Printf("UpdatedExpectedValues in %v over %d tries\n", time.Since(start), valueIters-1)
		if !changed {
			fmt.Println("No changes to ExpectedValues")
		}

		if err := save(mdp); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		start = time.Now()
		policyChanges := mdp.UpdatePolicy()
		fmt.Printf("UpdatePolicy in %v with %d changes\n", time.Since(start), policyChanges)

		if policyChanges == 0 {
			break
		}
	}
	fmt.Printf("Reached equilibrium for %d cutoff in %d iterations\n", cutoff, iter)
	return iter
}

// save saves the MDP to file.
func save(mdp *bot.MDP) error {
	// Write the new updated MDP back.
	start := time.Now()
	bytes, err := mdp.GobEncode()
	if err != nil {
		return fmt.Errorf("encode failed: %v", err)
	}
	if err := ioutil.WriteFile(*gobFile, []byte(bytes), 0644); err != nil {
		return fmt.Errorf("WriteFile failed: %v", err)
	}
	fmt.Printf("Updated file in %v\n", time.Since(start))
	return nil
}
