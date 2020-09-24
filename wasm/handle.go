package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var manager *GameManager

type GameManager struct {
	board    *game.Board
	tiles    []*game.Tile
	gridSize int
	tileCnt  int
}

func newGameManager(gridSize, tileCnt int) *GameManager {
	return &GameManager{
		board:    game.NewBoard(game.Vec{gridSize, gridSize}),
		gridSize: gridSize,
		tileCnt:  tileCnt,
	}
}

func sendTilesToTray() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})
}

func requestNewTile() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		websocketSendEmpty(game.MsgAddTile)
		return nil
	})
}
func reload() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})
}
func verify() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// TODO: add tiles
		m, _ := game.NewSocketMsg(game.MsgVerify, nil)
		websocketSend(m)
		return nil
	})
}

func addTile(t game.Tile) {}

func handleSocketMsg(t game.MessageType, data []byte) int {
	switch t {
	case game.MsgOK:
		if len(onOk) > 0 {
			onOk[0].Invoke()
			onOk = onOk[1:]
		}
	case game.MsgError:
	case game.MsgStart:
		if manager != nil {
			// TODO delete old manager
		}
		// TODO tie to actual game size
		manager = newGameManager(16, 16)
		err := json.Unmarshal(data, &manager.tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", manager.tiles)
	case game.MsgAddTile:
		var t game.Tile
		err := json.Unmarshal(data, &t)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		addTile(t)
		fmt.Println("Adding new tile:", t.Value)
	case game.MsgScore:

	case game.MsgGameStatus:
		var s game.GameStatus
		err := json.Unmarshal(data, &s)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("Game:", s.GameName)
	}
	return 0
}
