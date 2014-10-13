package main

import (
	"fmt"
)

var aTile = Tile{Value: "A", Points: pointValues["A"]}
var cTile = Tile{Value: "C", Points: pointValues["C"]}
var tTile = Tile{Value: "T", Points: pointValues["T"]}

func Example_PrintBoard() {
	b := makeBoard(2, 2, "A", "B", "", "C")
	fmt.Println(b.String())
	// Output: AB
	//  C
}

func Example_EmptyScore() {
	var testClient = &Client{maxScore: 7}
	b := makeBoard(2, 2, "", "", "", "")
	fmt.Println(b.scoreBoard(testClient).Score)
	// Output: 7
}

func Example_TallBoardScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile}}
	var testClient = &Client{maxScore: 5, tilesServed: 3, game: testGame}
	b := makeBoard(3, 1, "C", "A", "T")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {true 0}
}

func Example_LongBoardScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile}}
	var testClient = &Client{maxScore: 5, tilesServed: 3, game: testGame}
	b := makeBoard(1, 3, "C", "A", "T")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {true 0}
}

func Example_MultiWordPassingScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile, cTile, tTile}}
	var testClient = &Client{maxScore: 9, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {true 0}
}

func Example_MissingTilesScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile, cTile, tTile, cTile, aTile}}
	var testClient = &Client{maxScore: 13, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "A", "T", "", "C", "", "", "T", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {false 4}
}

func Example_MismatchedButValidTilesScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile, tTile}}
	var testClient = &Client{maxScore: 6, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "A", "T", "", "S", "", "", "", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {false 0}
}

func Example_MultiWordOneFailScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile, cTile, tTile}}
	var testClient = &Client{maxScore: 9, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "A", "T", "", "T", "", "", "C", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {false 9}
}

func Example_MultiComponentPassScore() {
	var testGame = &Game{tiles: Tiles{aTile, cTile, tTile, aTile, tTile}}
	var testClient = &Client{maxScore: 7, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "A", "T", "", "", "", "A", "T", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {false 2}
}

func Example_MultiComponentOneFailScore() {
	var testGame = &Game{tiles: Tiles{cTile, cTile, tTile, aTile, tTile}}
	var testClient = &Client{maxScore: 9, tilesServed: 5, game: testGame}
	b := makeBoard(3, 3, "C", "C", "T", "", "", "", "A", "T", "")
	fmt.Println(b.scoreBoard(testClient))
	// Output: {false 7}
}
