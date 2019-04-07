// This file is used for generating tetris/combo4/bot/scorer_gob.go
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"tetris/combo4/bot"
)

func main() {
	score := bot.GenScorer()
	bytes, err := score.GobEncode()
	if err != nil {
		fmt.Printf("encode failed: %v", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile("./scorer_gob.go", []byte(goString(bytes)), 0644); err != nil {
		fmt.Printf("WriteFile failed: %v", err)
		os.Exit(1)
	}
}

func goString(bytes []byte) string {
	var buf strings.Builder
	buf.WriteString(`package bot

// scorerGob is a precomputed GenScorer().GobEncode()
var scorerGob = []byte{`)
	for idx, b := range bytes {
		if idx == len(bytes)-1 {
			buf.WriteString(fmt.Sprintf("%d", b))
		} else {
			buf.WriteString(fmt.Sprintf("%d, ", b))
		}
	}
	buf.WriteString("}")
	return buf.String()
}
