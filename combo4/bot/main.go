// This package is to play 4 wide combos in NullpoMino.
package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"runtime"
	"tetris"
	"tetris/combo4"
	"tetris/combo4/policy"
	"time"

	"github.com/go-vgo/robotgo"
	kb "github.com/micmonay/keybd_event"
	"github.com/vova616/screenshot"
)

var (
	pressWait  = flag.Duration("press_delay", 25*time.Millisecond, "Time to wait between key presses.")
	lineWait   = flag.Duration("clear_delay", 0, "Time to wait for a line to clear.")
	policyFile = flag.String("policy_file", "policy_6preview.gob.gz", "Path the the gzip policy file. If empty-string, will compute an AI from scratch.")
)

const initialField = combo4.LeftI

// Co-ordinates to read the pixels of the last preview piece.
// Reads a square starting at lastPreviewPoint in the top left
// and moving readWith down and right.
var (
	lastPreviewPoint = image.Point{X: 1722, Y: 1030}
	readWidth        = 3
)

var actionKeys = map[tetris.Action]int{
	tetris.Left:      kb.VK_LEFT,
	tetris.Right:     kb.VK_RIGHT,
	tetris.SoftDrop:  kb.VK_DOWN,
	tetris.RotateCW:  kb.VK_UP,
	tetris.RotateCCW: kb.VK_Z,
	tetris.Hold:      kb.VK_C,
	tetris.HardDrop:  kb.VK_SPACE,
}

// Width of square to scan for pixel colors.
const scanWidth = 5

var colors = map[tetris.Piece]color.RGBA{
	tetris.Z: color.RGBA{R: 194, G: 27, B: 48},
	tetris.S: color.RGBA{R: 30, G: 205, B: 30},
	tetris.J: color.RGBA{R: 28, G: 49, B: 196},
	tetris.L: color.RGBA{R: 211, G: 121, B: 30},
	tetris.I: color.RGBA{R: 31, G: 191, B: 214},
	tetris.O: color.RGBA{R: 195, G: 181, B: 35},
	tetris.T: color.RGBA{R: 157, G: 21, B: 220},
}

func main() {

	// Parse the pieces from the args.
	args := os.Args
	if len(args) != 3 {
		log.Fatalf(`expected two args <first piece e.g "j"> <next known pieces e.g "ilostz">`)
	}
	first := tetris.PieceFromRune(rune((args[1])[0]))
	if first == tetris.EmptyPiece {
		log.Fatalf("piece must be one of %v.", tetris.NonemptyPieces)
	}
	next := tetris.SeqFromStr(args[2])
	for _, p := range next {
		if p == tetris.EmptyPiece {
			log.Fatalf("each piece must be one of %v.", tetris.NonemptyPieces)
		}
	}

	currPieceCh := make(chan tetris.Piece, len(next)+2)
	currPieceCh <- first
	for _, p := range next {
		currPieceCh <- p
	}

	fmt.Println("Loading AI...")
	moves, mActions := combo4.AllContinuousMoves()
	nfa := combo4.NewNFA(moves)
	var pol policy.Policy
	if *policyFile == "" {
		pol = policy.FromScorer(nfa, policy.NewNFAScorer(nfa, 7))
	} else {
		var err error
		pol, err = policyFromPath(*policyFile)
		if err != nil {
			fmt.Printf("failed to read policy from file: %v\n", err)
			os.Exit(1)
		}
	}

	keybond, err := newKeyBonding()
	if err != nil {
		fmt.Printf("newKeyBonding failed: %v", err)
		os.Exit(1)
	}

	fmt.Println("Middle click the mouse when you are ready for the bot to begin.")
	click := robotgo.AddEvent("center")
	if !click {
		fmt.Println("middle mouse button not clicked")
		os.Exit(1)
	}

	prevState := combo4.State{Field: initialField}

	policyInput := make(chan tetris.Piece, 1)

	for nextStatePtr := range policy.StartGame(pol, initialField, first, next, policyInput) {
		if nextStatePtr == nil {
			fmt.Println("No more combos!")
			return
		}
		nextState := *nextStatePtr

		currPiece := <-currPieceCh

		fmt.Printf("\nCurrent: %s\nHold: %s\nField:\n%s\n", currPiece, prevState.Hold, prevState.Field)

		toExecute := actions(mActions, prevState, nextState, currPiece)
		fmt.Println(toExecute)
		for _, a := range toExecute {
			k, ok := actionKeys[a]
			if !ok {
				panic(fmt.Sprintf("Unmapped tetris.Action = %v.\n", k))
			}
			if err := keyTap(keybond, k); err != nil {
				fmt.Printf("failed keyTap: %v\n", err)
				os.Exit(1)
			}
			time.Sleep(*pressWait)
		}

		// Wait for the line to clear.
		time.Sleep(*lineWait)

		// Read the next preview piece.
		nextPreview, err := lastPreviewPiece()
		if err != nil {
			fmt.Printf("getPiece failed: %v\n", err)
			os.Exit(1)
		}
		policyInput <- nextPreview
		currPieceCh <- nextPreview

		prevState = nextState
	}
}

func actions(mActions map[combo4.Move][]tetris.Action, prevState, nextState combo4.State, piece tetris.Piece) []tetris.Action {
	var actions []tetris.Action

	movePiece := piece
	if prevState.Hold != nextState.Hold {
		movePiece = prevState.Hold
		actions = append(actions, tetris.Hold)

		// No more actions are need if swapping from EmptyPiece.
		if prevState.Hold == tetris.EmptyPiece {
			return actions
		}
	}

	move := combo4.Move{
		Start: prevState.Field,
		End:   nextState.Field,
		Piece: movePiece,
	}
	moveActions, ok := mActions[move]
	if !ok {
		panic(fmt.Sprintf("no actions defined for move %+v", move))
	}
	actions = append(actions, moveActions...)
	return actions
}

func lastPreviewPiece() (tetris.Piece, error) {
	// Find the average color
	img, err := screenshot.CaptureRect(image.Rectangle{
		Min: image.Point{X: lastPreviewPoint.X - readWidth, Y: lastPreviewPoint.Y - readWidth},
		Max: image.Point{X: lastPreviewPoint.X + readWidth, Y: lastPreviewPoint.Y + readWidth},
	})
	if err != nil {
		return 0, fmt.Errorf("screenshot.Capture: %v", err)
	}
	var r, g, b int
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			c := img.RGBAAt(x, y)
			r += int(c.R)
			g += int(c.G)
			b += int(c.B)
		}
	}
	area := img.Bounds().Dx() * img.Bounds().Dy()
	r /= area
	g /= area
	b /= area
	var (
		minDistSq = -1
		piece     tetris.Piece
	)
	for p, c := range colors {
		distSq := (int(c.R)-r)*(int(c.R)-r) + (int(c.G)-g)*(int(c.G)-g) + (int(c.B)-b)*(int(c.B)-b)
		if minDistSq == -1 || minDistSq > distSq {
			minDistSq = distSq
			piece = p
		}
	}
	return piece, nil
}

func newKeyBonding() (*kb.KeyBonding, error) {
	kb, err := kb.NewKeyBonding()
	if err != nil {
		return nil, fmt.Errorf("NewKeyBonding: %v", err)
	}
	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}
	return &kb, nil
}

func keyTap(keybnd *kb.KeyBonding, key int) error {
	keybnd.Clear()
	keybnd.SetKeys(key)
	return keybnd.Launching()
}

func policyFromPath(path string) (policy.Policy, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	gz, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader: %v", err)
	}
	defer gz.Close()

	if _, err := io.Copy(&buf, gz); err != nil {
		return nil, fmt.Errorf("read file contents failed: %v", err)
	}

	mdpPol := &policy.MDPPolicy{}
	if err := mdpPol.GobDecode(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("GobDecode failed: %v", err)
	}
	return mdpPol, nil
}
