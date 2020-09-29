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
	Valid       []Vec  // Tiles which were part of valid words.
	Invalid     []Vec  // Tiles which were part of no valid words.
	Unconnected []Vec  // Tiles not part of the best scoring component.
	Words       []Word // Words found in the dictionary.
	Nonwords    []Word // Words not found in the dictionary.
}

func (mgr *GameManager) joinGame() {
	m, _ := msg.NewSocketData(msg.JoinGame, "NAME")
	mgr.websocketSend(m)
}

func (mgr *GameManager) requestNewTile() {
	mgr.websocketSendEmpty(msg.AddTile)
}

func (mgr *GameManager) newGame() {
	if mgr.state == StateNoGame {
		m, _ := msg.NewSocketData(msg.JoinGame, "NAME")
		mgr.websocketSend(m)
	} else {
		mgr.websocketSendEmpty(msg.RoundReady)
	}
}

func (mgr *GameManager) verify() {
	// TODO: send only tiles instead of entire board
	m, _ := msg.NewSocketData(msg.Verify, mgr.board.Grid)
	mgr.websocketSend(m)
}

func (mgr *GameManager) handleSocketMsg(t msg.Type, data []byte) int {
	switch t {
	case msg.PlayerJoined:
		mgr.websocketSendEmpty(msg.RoundReady)
	case msg.Error:
	case msg.Start:
		// TODO tie to actual game size
		mgr.Reset()
		var tiles []*Tile
		err := json.Unmarshal(data, &tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", tiles)
		for _, tile := range tiles {
			tile.mgr = mgr
			mgr.tiles = append(mgr.tiles, tile)
			tile.sendToTray()
		}
		mgr.listens.NewGame()
		mgr.state = StatePlaying
		EnableGameButtons()
		mgr.draw()
	case msg.AddTile:
		var tile *Tile
		err := json.Unmarshal(data, &tile)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		tile.mgr = mgr
		mgr.tiles = append(mgr.tiles, tile)
		tile.sendToTray()
		mgr.draw()
	case msg.Result:
		var score Score
		err := json.Unmarshal(data, &score)
		if err != nil {
			fmt.Println("Error reading score:", err)
			return 1
		}
		mgr.listens.EndGame()
		DisableGameButtons()
		mgr.unhighlight()
		mgr.markInvalidAndUnusedTiles(score.Invalid, score.Unconnected, score.Nonwords)
		mgr.draw()
		mgr.state = StateGameOver
	case msg.Invalid:
		var score Score
		err := json.Unmarshal(data, &score)
		if err != nil {
			fmt.Println("Error reading score:", err)
			return 1
		}
		mgr.markInvalidAndUnusedTiles(score.Invalid, score.Unconnected, score.Nonwords)
		mgr.draw()
	case msg.SendBoard:
		m, _ := msg.NewSocketData(msg.SendBoard, mgr.board.Grid)
		mgr.websocketSend(m)
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
