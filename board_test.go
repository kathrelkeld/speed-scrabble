package main

import (
	"fmt"
	"testing"
)

func testError(t *testing.T, got, expected interface{}, name string) {
	t.Errorf("%s: Got %v; Expected %v", name, got, expected)
}

var aTile = Tile{Value: "A", Points: pointValues["A"]}
var cTile = Tile{Value: "C", Points: pointValues["C"]}
var tTile = Tile{Value: "T", Points: pointValues["T"]}

func TestPrintBoard(t *testing.T) {
	e := "AB\n C"
	b := makeBoard(2, 2, "A", "B", "", "C")
	if g := b.String(); g != "AB\n C\n" {
		testError(t, g, e, "PrintBoard")
	}
}

var scoreTests = []struct {
	b     Board
	tiles []string
	e     Score
	name  string
}{
	{makeBoard(2, 2, "", "", "", ""), []string{"A", "C", "C"},
		Score{false, 7}, "EmptyBoard"},
	{makeBoard(3, 1, "C", "A", "T"), []string{"C", "A", "T"},
		Score{true, 0}, "TallBoard"},
	{makeBoard(1, 3, "C", "A", "T"), []string{"C", "A", "T"},
		Score{true, 0}, "LongBoard"},
	{makeBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		[]string{"C", "A", "T", "C", "T"}, Score{true, 0}, "MutliWordPassing"},
	{makeBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", ""),
		[]string{"C", "A", "T", "C", "T", "C", "A"},
		Score{false, 4}, "MissingTilesButPassing"},
	{makeBoard(3, 3, "C", "A", "T", "", "", "A", "", "", "N"),
		[]string{"C", "A", "T", "T"}, Score{false, 6}, "MismatchedScore"},
	{makeBoard(3, 3, "C", "A", "T", "", "", "S", "", "", ""),
		[]string{"C", "A", "T", "T"}, Score{false, 6}, "MismatchedTilesButValid"},
	{makeBoard(3, 3, "C", "A", "T", "", "", "A", "", "", "D"),
		[]string{"C", "A", "T", "C"}, Score{false, 8}, "TooManyBoardTiles"},
	{makeBoard(3, 3, "C", "A", "T", "", "E", "", "", "E", ""),
		[]string{"C", "A", "T", "E", "E"}, Score{false, 7}, "MultiWordOneFails"},
	{makeBoard(3, 3, "C", "A", "T", "", "", "", "", "A", "T"),
		[]string{"C", "A", "T", "A", "T"}, Score{false, 2}, "MultiComponentAllPass"},
	{makeBoard(3, 3, "C", "C", "T", "", "", "", "", "A", "T"),
		[]string{"C", "A", "T", "C", "T"}, Score{false, 7}, "MultiComponentOneFails"},
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
		var testGame = &Game{tiles: tiles}
		var testClient = &Client{maxScore: maxScore, tilesServed: len(tiles),
			game: testGame}
		if g := input.b.scoreBoard(testClient); g != input.e {
			testError(t, g, input.e, "Score of "+input.name)
		}
	}
}
