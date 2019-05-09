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
	"math"
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

var actionKeys = map[tetris.Action]int{
	tetris.Left:      kb.VK_LEFT,
	tetris.Right:     kb.VK_RIGHT,
	tetris.SoftDrop:  kb.VK_DOWN,
	tetris.RotateCW:  kb.VK_UP,
	tetris.RotateCCW: kb.VK_Z,
	tetris.Hold:      kb.VK_C,
	tetris.HardDrop:  kb.VK_SPACE,
}

// Co-ordinates to read the pixels of the preview pieces.
// These defaults are how NullpoMino opens on a 4K screen.
var (
	// This assumes the initialField is LeftI.
	initialCurrPoint = image.Point{X: 1500, Y: 1400}

	previewPoints = []image.Point{
		{X: 1500, Y: 782},
		{X: 1620, Y: 790},
		{X: 1700, Y: 790},
		{X: 1725, Y: 870},
		{X: 1725, Y: 950},
		{X: 1725, Y: 1030},
	}

	holdPoint = image.Point{X: 1370, Y: 790}

	// Reads a square starting at the points in the top left
	// and moving readWith down and right.
	readWidth = 3
)

var colors = map[tetris.Piece]color.RGBA{
	// Assuming no/black background.
	tetris.EmptyPiece: color.RGBA{R: 0, G: 0, B: 0},

	tetris.Z: color.RGBA{R: 194, G: 27, B: 48},
	tetris.S: color.RGBA{R: 30, G: 205, B: 30},
	tetris.J: color.RGBA{R: 28, G: 49, B: 196},
	tetris.L: color.RGBA{R: 211, G: 121, B: 30},
	tetris.I: color.RGBA{R: 31, G: 191, B: 214},
	tetris.O: color.RGBA{R: 195, G: 181, B: 35},
	tetris.T: color.RGBA{R: 157, G: 21, B: 220},
}

var moves, mActions = combo4.AllContinuousMoves()

func main() {
	fmt.Println("Loading AI...")
	var pol policy.Policy
	if *policyFile == "" {
		nfa := combo4.NewNFA(moves)
		pol = policy.FromScorer(nfa, policy.NewNFAScorer(nfa, 7))
	} else {
		var err error
		pol, err = policyFromPath(*policyFile)
		if err != nil {
			log.Fatalf("failed to read policy from file: %v\n", err)
		}
	}

	keybond, err := newKeyBonding()
	if err != nil {
		log.Fatalf("newKeyBonding failed: %v", err)
	}

	for {
		playGame(pol, keybond)
	}
}

func playGame(pol policy.Policy, keybond *kb.KeyBonding) {
	fmt.Println("Middle click the mouse when you are ready for the bot to begin.")
	click := robotgo.AddEvent("center")
	if !click {
		log.Fatal("middle mouse button not clicked")
	}

	// Read the pieces from the screen.
	piecePnts := append([]image.Point{initialCurrPoint}, previewPoints...)
	var initialPieces []tetris.Piece
	for _, pnt := range piecePnts {
		piece := pieceAt(pnt)
		if piece == tetris.EmptyPiece {
			log.Fatalf("got EmptyPiece piece at %v.", pnt)
		}
		initialPieces = append(initialPieces, piece)
	}
	currPieceCh := make(chan tetris.Piece, len(initialPieces)+1)
	for _, p := range initialPieces {
		currPieceCh <- p
	}
	fmt.Printf("First piece: %v\n", initialPieces[0])
	fmt.Printf("Preview: %v\n", initialPieces[1:])

	var (
		prevState   = combo4.State{Field: initialField}
		policyInput = make(chan tetris.Piece, 1)
	)
	for nextStatePtr := range policy.StartGame(pol, initialField, initialPieces[0], initialPieces[1:], policyInput) {
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
			keyTap(keybond, k)
			time.Sleep(*pressWait)
		}

		time.Sleep(*lineWait)

		// Read the new last preview piece.
		nextPreview := pieceAt(previewPoints[len(previewPoints)-1])
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

// pieceAt returns the piece at a point or exits the program.
func pieceAt(point image.Point) tetris.Piece {
	// Find the average color
	img, err := screenshot.CaptureRect(image.Rectangle{
		Min: image.Point{X: point.X - readWidth, Y: point.Y - readWidth},
		Max: image.Point{X: point.X + readWidth, Y: point.Y + readWidth},
	})
	if err != nil {
		log.Fatalf("failed to read piece at %v: %v", point, err)
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
		minDistSq = math.MaxInt32
		piece     tetris.Piece
	)
	for p, c := range colors {
		distSq := (int(c.R)-r)*(int(c.R)-r) + (int(c.G)-g)*(int(c.G)-g) + (int(c.B)-b)*(int(c.B)-b)
		if minDistSq <= distSq {
			continue
		}
		minDistSq = distSq
		piece = p
	}
	return piece
}

func newKeyBonding() (*kb.KeyBonding, error) {
	kb, err := kb.NewKeyBonding()
	if err != nil {
		return nil, fmt.Errorf("NewKeyBonding: %v", err)
	}
	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		fmt.Println("Creating fake keyboard...")
		time.Sleep(2 * time.Second)
	}
	return &kb, nil
}

// keyTap presses a key or exits.
func keyTap(keybnd *kb.KeyBonding, key int) {
	keybnd.Clear()
	keybnd.SetKeys(key)
	if err := keybnd.Launching(); err != nil {
		log.Fatalf("key press failed: %v", err)
	}
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
