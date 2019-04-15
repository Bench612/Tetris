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
	gobFile     = flag.String("gob_file", "mdp5.gob", "The path to a binary file of the MDP gob encoding")
	iterations  = flag.Int("iterations", 5, "The number of times to iterate")
	fromScratch = flag.Bool("from_scratch", false, "If set to true, does not read the MDP from file but creates a new one")
)

func main() {
	flag.Parse()

	var mdp *bot.MDP
	if *fromScratch {
		var err error
		start := time.Now()
		mdp, err = bot.NewMDP(5)
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

	for iter := 0; iter < *iterations; iter++ {
		// Now update the MDP iteratively.
		start := time.Now()
		for i := 0; i < 5; i++ {
			if updated := mdp.UpdateExpectedValues(); updated == 0 {
				if i == 0 {
					fmt.Println("No significant changes to ExpectedValues")
					break
				}
			}
		}
		fmt.Printf("UpdatedExpectedValues in %v\n", time.Since(start))

		start = time.Now()
		policyChanges := mdp.UpdatePolicy()
		fmt.Printf("UpdatePolicy in %v with %d changes\n", time.Since(start), policyChanges)

		if iter%10 == 0 {
			if err := Save(mdp); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
	if err := Save(mdp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Save saves the MDP to file.
func Save(mdp *bot.MDP) error {
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
