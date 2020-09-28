package main

import (
	"encoding/json"
	"fmt"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

type TileSet map[Vec]*Tile

type Word struct {
	Start Vec
	End   Vec
	Value string
}

type Score struct {
	Win         bool   // Whether the board ends the game or not.
	Pts         int    // The numerical score (lower is better).
	Valid       []Vec  // Tiles which were part of no valid words.
	Invalid     []Vec  // Tiles which were part of no valid words.
	Unconnected []Vec  // Tiles not part of the best scoring component.
	Words       []Word // Words found in the dictionary.
	Nonwords    []Word // Words not found in the dictionary.
}

func joinGame() {
	m, _ := msg.NewSocketData(msg.JoinGame, "NAME")
	websocketSend(m)
}

func requestNewTile() {
	websocketSendEmpty(msg.AddTile)
}

func newGame() {
	websocketSendEmpty(msg.RoundReady)
}

func verify() {
	// TODO: send only tiles instead of entire board
	m, _ := msg.NewSocketData(msg.Verify, mgr.board.Grid)
	websocketSend(m)
}

func handleSocketMsg(t msg.Type, data []byte) int {
	switch t {
	case msg.PlayerJoined:
		websocketSendEmpty(msg.RoundReady)
	case msg.Error:
	case msg.Start:
		// TODO tie to actual game size
		resetGameManager(Vec{16, 16}, 16)
		var tiles []*Tile
		err := json.Unmarshal(data, &tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", tiles)
		for _, tile := range tiles {
			mgr.tiles = append(mgr.tiles, tile)
			tile.sendToTray()
		}
		draw()
	case msg.AddTile:
		var tile *Tile
		err := json.Unmarshal(data, &tile)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		mgr.tiles = append(mgr.tiles, tile)
		tile.sendToTray()
		draw()
	case msg.Score:
		var score Score
		err := json.Unmarshal(data, &score)
		if err != nil {
			fmt.Println("Error reading score:", err)
			return 1
		}
	case msg.Invalid:
		var score Score
		err := json.Unmarshal(data, &score)
		if err != nil {
			fmt.Println("Error reading score:", err)
			return 1
		}
		markInvalidTiles(score.Invalid)
		draw()
	case msg.GameInfo:
		var s msg.GameInfoData
		err := json.Unmarshal(data, &s)
		if err != nil {
			fmt.Println("Error reading game info:", err)
			return 1
		}
		fmt.Println("Game:", s.GameName)
	}
	return 0
}
