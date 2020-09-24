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
	for i := 0; i < x; i++ {
		b = append(b, letters[i*y:(i+1)*y])
	}
	return b
}

func makeTestVecSet(vs ...Vec) VecSet {
	result := make(VecSet)
	for _, v := range vs {
		result.insert(v)
	}
	return result
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
	name  string
	board Board
	tiles []string
	score Score
}{
	{name: "Empty board",
		board: makeTestBoard(2, 2, "", "", "", ""),
		tiles: []string{"A", "C", "C"},
		score: Score{false, 7, makeTestVecSet()}},
	{name: "Tall board",
		board: makeTestBoard(3, 1, "C", "A", "T"),
		tiles: []string{"C", "A", "T"},
		score: Score{true, 0, makeTestVecSet()}},
	{name: "Long board",
		board: makeTestBoard(1, 3, "C", "A", "T"),
		tiles: []string{"C", "A", "T"},
		score: Score{true, 0, makeTestVecSet()}},
	{name: "One component, multi word, all used, all valid",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		tiles: []string{"C", "A", "T", "C", "T"},
		score: Score{true, 0, makeTestVecSet()}},
	{name: "Missing tiles but otherwise passing",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		tiles: []string{"C", "A", "T", "C", "T", "C", "A"},
		score: Score{false, 4, makeTestVecSet()}},
	{name: "One component, all of the words fail",
		board: makeTestBoard(3, 3, "C", "A", "O", "", "E", "", "", "E", ""),
		tiles: []string{"C", "A", "O", "E", "E"},
		score: Score{false, 7,
			makeTestVecSet(Vec{0, 0}, Vec{0, 1}, Vec{0, 2}, Vec{1, 1}, Vec{2, 1})}},
	{name: "One component, one of the words fails but falls off cleanly",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "E", "", "", "E", ""),
		tiles: []string{"C", "A", "T", "E", "E"},
		score: Score{false, 2, makeTestVecSet(Vec{0, 1}, Vec{1, 1}, Vec{2, 1})}},
	{name: "One component, failing word divides it into two components",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "E", "", "M", "E", ""),
		tiles: []string{"C", "A", "T", "E", "E", "M"},
		score: Score{false, 5, makeTestVecSet(Vec{0, 1}, Vec{1, 1}, Vec{2, 1})}},
	{name: "Unremovable tiles (one set)",
		board: makeTestBoard(3, 3, "Q", "I", "", "A", "S", "", "", "", ""),
		tiles: []string{"Q", "I", "A", "S"},
		score: Score{false, 1, makeTestVecSet(Vec{0, 0}, Vec{1, 0})}},
	{name: "Unremovable tiles (multiple sets)",
		board: makeTestBoard(3, 3, "A", "P", "T", "F", "A", "N", "", "", ""),
		tiles: []string{"A", "P", "T", "F", "A", "N"},
		score: Score{false, 2,
			makeTestVecSet(Vec{0, 0}, Vec{1, 0}, Vec{0, 2}, Vec{1, 2})}},
	{name: "Multi component, all pass",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "", "", "", "A", "T"),
		tiles: []string{"C", "A", "T", "A", "T"},
		score: Score{false, 2, makeTestVecSet()}},
	{name: "Multi component, one invalid",
		board: makeTestBoard(3, 3, "C", "C", "T", "", "", "", "", "A", "T"),
		tiles: []string{"C", "C", "T", "A", "T"},
		score: Score{false, 7, makeTestVecSet(Vec{0, 0}, Vec{0, 1}, Vec{0, 2})}},
	{name: "Multi component, one invalid",
		board: makeTestBoard(3, 3, "C", "C", "T", "", "", "", "", "A", "T"),
		tiles: []string{"C", "C", "T", "A", "T"},
		score: Score{false, 7, makeTestVecSet(Vec{0, 0}, Vec{0, 1}, Vec{0, 2})}},
	{name: "Multi component, one partially invalid",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "", "S", "A", "T", ""),
		tiles: []string{"C", "A", "T", "S", "A", "T"},
		score: Score{false, 3, makeTestVecSet(Vec{0, 2}, Vec{1, 2})}},
	{name: "Too many board tiles, passing tiles",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "", "O", "", "", "O"),
		tiles: []string{"C", "A", "T"},
		score: Score{false, 5, makeTestVecSet()}},
	{name: "Too many board tiles, failing tiles",
		board: makeTestBoard(3, 3, "C", "A", "S", "", "", "O", "", "", "O"),
		tiles: []string{"C", "A", "O"},
		score: Score{false, 5,
			makeTestVecSet(Vec{0, 0}, Vec{0, 1}, Vec{0, 2}, Vec{1, 2}, Vec{2, 2})}},
	{name: "Mismatched tile values, passing tiles",
		board: makeTestBoard(3, 3, "C", "A", "T", "", "", "O", "", "", "O"),
		tiles: []string{"C", "A", "T", "A", "A"},
		score: Score{false, 7, makeTestVecSet()}},
	{name: "Mismatched tile values, failing tiles",
		board: makeTestBoard(3, 3, "C", "A", "S", "", "", "O", "", "", "O"),
		tiles: []string{"C", "A", "T", "A", "A"},
		score: Score{false, 7,
			makeTestVecSet(Vec{0, 0}, Vec{0, 1}, Vec{0, 2}, Vec{1, 2}, Vec{2, 2})}},
}

func TestVariousBoardScores(t *testing.T) {
	for _, input := range scoreTests {
		fmt.Println("Starting test:", input.name)
		tiles := Tiles{}
		maxScore := 0
		for _, elt := range input.tiles {
			tiles = append(tiles, Tile{elt, pointValues[elt]})
			maxScore += pointValues[elt]
		}
		testClient := makeNewClient()
		testClient.newTiles(tiles)
		s := input.board.scoreBoard(testClient)
		if !reflect.DeepEqual(s, input.score) {
			testError(t, s, input.score, input.name)
		}
	}
}
