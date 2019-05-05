// This package is used for using the bot to play 4 wide combos through the terminal.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"tetris"
	"tetris/combo4"
	"tetris/combo4/bot"
)

func main() {
	args := os.Args
	if len(args) != 3 {
		log.Fatalf("expected two args <first piece> <next known pieces>")
	}

	first := tetris.PieceFromRune(rune((args[1])[0]))
	if first == tetris.EmptyPiece {
		log.Fatalf("piece must be one of %v", tetris.NonemptyPieces)
	}

	next := tetris.SeqFromStr(args[2])
	for _, p := range next {
		if p == tetris.EmptyPiece {
			log.Fatalf("each piece must be one of %v", tetris.NonemptyPieces)
		}
	}

	moves, _ := combo4.AllContinuousMoves()
	nfa := combo4.NewNFA(moves)

	input := make(chan tetris.Piece, 1)
	output := bot.StartGame(bot.PolicyFromScorer(nfa, bot.NewNFAScorer(nfa, 8)), combo4.LeftI, first, next, input)

	reader := bufio.NewReader(os.Stdin)
	for state := range output {
		if state == nil {
			fmt.Println("No more combos!")
			return
		}
		fmt.Println(state)

		for {
			fmt.Printf("Newest next piece (q to quit): ")
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("failed to read from stdin: %v", err)
			}
			if strings.HasPrefix(text, "q") {
				fmt.Println("goodbye!")
				return
			}
			p := tetris.PieceFromRune(rune(text[0]))

			if p == tetris.EmptyPiece {
				fmt.Printf("input=%q must be one of %v\n", text[0], tetris.NonemptyPieces)
			} else {
				input <- p
				break
			}
		}
	}
}
