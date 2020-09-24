package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var manager *GameManager

type GameManager struct {
	board     *game.Board
	tiles     []*TileLoc
	boardSize game.Vec
	traySize  game.Vec
	tileCnt   int
	tileSize  game.Vec
}

func newGameManager(boardSize game.Vec, tileCnt int) *GameManager {
	return &GameManager{
		board:     game.NewBoard(boardSize),
		boardSize: boardSize,
		tileCnt:   tileCnt,
		traySize:  game.Vec{tileCnt, 1},
		tileSize:  game.Vec{25, 25},
	}
}

type TileLoc struct {
	value  string
	inPlay bool
	moving bool
	loc    game.Vec
}

func newTileLoc(v string) *TileLoc {
	return &TileLoc{
		value:  v,
		inPlay: false,
		moving: false,
		loc:    game.Vec{-1, -1},
	}
}

func (t *TileLoc) collides(x, y int) bool {
	return ((x < t.loc.X+manager.tileSize.X) && (x > t.loc.X) &&
		(y < t.loc.Y+manager.tileSize.Y) && (y > t.loc.Y))
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
		manager = newGameManager(game.Vec{16, 16}, 16)
		var tiles []*game.Tile
		err := json.Unmarshal(data, &tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", tiles)
		for _, tile := range tiles {
			manager.tiles = append(manager.tiles, newTileLoc(tile.Value))
		}
		draw()
	case game.MsgAddTile:
		var tile game.Tile
		err := json.Unmarshal(data, &t)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		manager.tiles = append(manager.tiles, newTileLoc(tile.Value))
		fmt.Println("Adding new tile:", tile.Value)
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
