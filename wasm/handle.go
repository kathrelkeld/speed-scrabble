package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var mgr *GameManager // initiated in page setup.
func joinGame() {
	m, _ := msg.NewSocketData(msg.JoinGame, "NAME")
	websocketSend(m)
}

func requestNewTile() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		websocketSendEmpty(msg.AddTile)
		return nil
	})
}
func newGame() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		websocketSendEmpty(msg.RoundReady)
		return nil
	})
}
func verify() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// TODO: add tiles
		m, _ := msg.NewSocketData(msg.Verify, mgr.board.Grid)
		websocketSend(m)
		return nil
	})
}

func handleSocketMsg(t msg.Type, data []byte) int {
	switch t {
	case msg.PlayerJoined:
		websocketSendEmpty(msg.RoundReady)
	case msg.Error:
	case msg.RoundReady:
		// Game is ready.  Need to reply with msg.Start player is ready.
	case msg.Start:
		if mgr != nil {
			// TODO delete old manager
		}
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

	case msg.Invalid:
		var invalid []Vec
		err := json.Unmarshal(data, &invalid)
		if err != nil {
			fmt.Println("Error reading invalid tiles:", err)
			return 1
		}
		markInvalidTiles(invalid)
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
