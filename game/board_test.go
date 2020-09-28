package game

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func makeTestBoard(x, y int, letters ...string) Board {
	b := Board{}
	if len(letters) != x*y {
		log.Println("Letters count did not match given dimensions!")
		return b
	}
	for j := 0; j < y; j++ {
		row := []*Tile{}
		for i := 0; i < x; i++ {
			if l := letters[j*x+i]; l != "" {
				row = append(row, &Tile{Value: l, Points: pointValues[l]})
			} else {
				row = append(row, nil)
			}
		}
		b = append(b, row)
	}
	return b
}

func testError(t *testing.T, got, expected interface{}, name string) {
	t.Errorf("%s: Got %v; Expected %v", name, got, expected)
}

func TestPrintBoard(t *testing.T) {
	e := "AB\n C"
	b := makeTestBoard(2, 2, "A", "B", "", "C")
	if g := b.String(); g != "AB\n C\n" {
		testError(t, g, e, "PrintBoard")
	}
}

var scoreTests = []struct {
	name    string
	board   Board
	tiles   []string
	isWin   bool
	score   int
	unconn  []Vec
	invalid []Vec
	words   []string
	nonw    []string
}{
	{name: "Empty board",
		board: makeTestBoard(2, 2, "", "", "", ""),
		tiles: []string{"A", "C", "C"},
		score: 7,
	},
	{name: "Tall board",
		board: makeTestBoard(3, 1, "C", "A", "T"),
		tiles: []string{"C", "A", "T"},
		isWin: true,
		words: []string{"CAT"},
	},
	{name: "Long board",
		board: makeTestBoard(1, 3, "C", "A", "T"),
		tiles: []string{"C", "A", "T"},
		isWin: true,
		words: []string{"CAT"},
	},
	{name: "One component, multi word, all used, all valid",
		board:   makeTestBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		tiles:   []string{"C", "A", "T", "C", "T"},
		isWin:   true,
		invalid: []Vec{},
		words:   []string{"CAT", "ACT"},
	},
	{name: "Missing tiles but otherwise passing",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		tiles: []string{"C", "A", "T", "C", "T", "C", "A"},
		score: 4,
		words: []string{"CAT", "ACT"},
	},
	{name: "One component, all of the words fail",
		board:   makeTestBoard(3, 3, "C", "A", "O", "", "E", "", "", "E", ""),
		tiles:   []string{"C", "A", "O", "E", "E"},
		score:   7,
		invalid: []Vec{Vec{0, 0}, Vec{1, 0}, Vec{2, 0}, Vec{1, 1}, Vec{1, 2}},
		nonw:    []string{"CAO", "AEE"},
	},
	{name: "One component, one of the words fails but falls off cleanly",
		board:   makeTestBoard(3, 3, "C", "A", "T", "", "E", "", "", "E", ""),
		tiles:   []string{"C", "A", "T", "E", "E"},
		score:   2,
		invalid: []Vec{Vec{1, 1}, Vec{1, 2}},
		words:   []string{"CAT"},
		nonw:    []string{"AEE"},
	},
	{name: "One component, failing word divides it into two components",
		board:   makeTestBoard(3, 3, "C", "A", "T", "", "E", "", "M", "E", ""),
		tiles:   []string{"C", "A", "T", "E", "E", "M"},
		score:   5,
		unconn:  []Vec{Vec{0, 2}, Vec{1, 2}},
		invalid: []Vec{Vec{1, 1}},
		words:   []string{"CAT", "ME"},
		nonw:    []string{"AEE"},
	},
	{name: "Unremovable tiles (one set)",
		board:   makeTestBoard(3, 3, "Q", "I", "", "A", "S", "", "", "", ""),
		tiles:   []string{"Q", "I", "A", "S"},
		score:   1,
		invalid: []Vec{Vec{0, 1}},
		words:   []string{"QI", "AS", "IS"},
		nonw:    []string{"QA"},
	},
	{name: "Unremovable tiles (multiple sets)",
		board:   makeTestBoard(3, 3, "A", "P", "T", "F", "A", "N", "", "", ""),
		tiles:   []string{"A", "P", "T", "F", "A", "N"},
		score:   2,
		invalid: []Vec{Vec{0, 0}, Vec{2, 0}},
		words:   []string{"APT", "FAN", "PA"},
		nonw:    []string{"AF", "TN"},
	},
	{name: "Multi component, all pass",
		board:  makeTestBoard(3, 3, "C", "A", "T", "", "", "", "", "A", "T"),
		tiles:  []string{"C", "A", "T", "A", "T"},
		score:  2,
		unconn: []Vec{Vec{1, 2}, Vec{2, 2}},
		words:  []string{"CAT", "AT"},
	},
	{name: "Multi component, one invalid",
		board:   makeTestBoard(3, 3, "C", "C", "T", "", "", "", "", "A", "T"),
		tiles:   []string{"C", "C", "T", "A", "T"},
		score:   7,
		invalid: []Vec{Vec{0, 0}, Vec{1, 0}, Vec{2, 0}},
		words:   []string{"AT"},
		nonw:    []string{"CCT"},
	},
	{name: "Multi component, one partially invalid",
		board:   makeTestBoard(3, 3, "C", "A", "T", "", "", "S", "A", "T", ""),
		tiles:   []string{"C", "A", "T", "S", "A", "T"},
		score:   3,
		invalid: []Vec{Vec{2, 1}},
		unconn:  []Vec{Vec{0, 2}, Vec{1, 2}},
		words:   []string{"CAT", "AT"},
		nonw:    []string{"TS"},
	},
	{name: "Too many board tiles, passing tiles",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "", "O", "", "", "O"),
		tiles: []string{"C", "A", "T"},
		score: 5,
		words: []string{"CAT", "TOO"},
	},
	{name: "Too many board tiles, failing tiles",
		board:   makeTestBoard(3, 3, "C", "A", "S", "", "", "O", "", "", "O"),
		tiles:   []string{"C", "A", "O"},
		score:   5,
		invalid: []Vec{Vec{0, 0}, Vec{1, 0}, Vec{2, 0}, Vec{2, 1}, Vec{2, 2}},
		nonw:    []string{"CAS", "SOO"},
	},
	{name: "Mismatched tile values, passing tiles",
		board:   makeTestBoard(3, 3, "C", "A", "T", "", "", "O", "", "", "O"),
		tiles:   []string{"C", "A", "T", "A", "A"},
		score:   7,
		invalid: []Vec{},
		words:   []string{"CAT", "TOO"},
	},
	{name: "Mismatched tile values, failing tiles",
		board:   makeTestBoard(3, 3, "C", "A", "S", "", "", "O", "", "", "O"),
		tiles:   []string{"C", "A", "T", "A", "A"},
		score:   7,
		invalid: []Vec{Vec{0, 0}, Vec{1, 0}, Vec{2, 0}, Vec{2, 1}, Vec{2, 2}},
		nonw:    []string{"CAS", "SOO"},
	},
}

func cmp(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func cmpTileSetAndList(s TileSet, vs []Vec) bool {
	if len(s) != len(vs) {
		return false
	}
	for _, elt := range vs {
		if !s.contains(elt) {
			return false
		}
	}
	return true
}

func cmpWordsAndList(ws []Word, ls []string) bool {
	if len(ws) != len(ls) {
		return false
	}
	for _, w := range ws {
		present := false
		for _, s := range ls {
			if w.Value == s {
				present = true
				break
			}
		}
		if present == false {
			return false
		}
	}
	for _, s := range ls {
		present := false
		for _, w := range ws {
			if w.Value == s {
				present = true
				break
			}
		}
		if present == false {
			return false
		}
	}
	return true
}

func TestVariousBoardScores(t *testing.T) {
	for _, input := range scoreTests {
		fmt.Println("Starting test:", input.name)
		tiles := []Tile{}
		maxScore := 0
		for _, elt := range input.tiles {
			tiles = append(tiles, Tile{elt, pointValues[elt]})
			maxScore += pointValues[elt]
		}
		s := input.board.scoreBoard(tiles)
		if !cmp(s.Pts, input.score) {
			testError(t, s.Pts, input.score, input.name+" - score")
		}
		if !cmp(s.Win, input.isWin) {
			testError(t, s.Win, input.isWin, input.name+" - win")
		}
		if !cmpTileSetAndList(s.Invalid, input.invalid) {
			testError(t, s.Invalid, input.invalid, input.name+" - invalid")
		}
		if !cmpTileSetAndList(s.Unconnected, input.unconn) {
			testError(t, s.Unconnected, input.unconn, input.name+" - unconnected")
		}
		if !cmpWordsAndList(s.Words, input.words) {
			testError(t, s.Words, input.words, input.name+" - words")
		}
		if !cmpWordsAndList(s.Nonwords, input.nonw) {
			testError(t, s.Nonwords, input.nonw, input.name+" - nonwords")
		}
	}
}
