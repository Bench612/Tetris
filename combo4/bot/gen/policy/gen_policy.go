package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"tetris/combo4/bot"
	"time"
)

var (
	mdpFile    = flag.String("mdp_file", "mdp5.gob", "The path to a binary file of the MDP gob encoding")
	policyFile = flag.String("policy_file", "mdp_policy5.gob", "The path to write the binary file of the MDPPolicy")
)

func main() {
	flag.Parse()

	start := time.Now()
	bytes, err := ioutil.ReadFile(*mdpFile)
	if err != nil {
		fmt.Printf("failed to read file at %q: %v\n", *mdpFile, err)
		os.Exit(1)
	}
	mdp := &bot.MDP{}
	if err := mdp.GobDecode(bytes); err != nil {
		fmt.Printf("GobDecode failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got initial MDP in %v\n", time.Since(start))

	// Release resouces
	bytes = nil

	policy := mdp.Policy()

	// Release resources.
	mdp = nil

	start = time.Now()
	bytes, err = policy.GobEncode()
	if err != nil {
		fmt.Printf("encode failed: %v", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(*policyFile, []byte(bytes), 0644); err != nil {
		fmt.Printf("WriteFile failed: %v", err)
		os.Exit(1)
	}
	log.Printf("Updated file in %v\n", time.Since(start))
}
