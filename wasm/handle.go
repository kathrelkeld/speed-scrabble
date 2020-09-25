package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var manager *GameManager

type sizeV game.Vec
type gridLoc game.Vec
type canvasLoc game.Vec
type board [][]*TileLoc

func newBoard(size sizeV) *board {
	var b board
	for i := 0; i < size.Y; i++ {
		b = append(b, []*TileLoc{})
		for j := 0; j < size.X; j++ {
			b[i] = append(b[i], nil)
		}
	}
	return &b
}

type GameManager struct {
	board     *board
	tray      *board
	tiles     []*TileLoc
	boardSize sizeV
	traySize  sizeV
	tileCnt   int
	tileSize  sizeV
}

func newGameManager(boardSize sizeV, tileCnt int) *GameManager {
	traySize := sizeV{tileCnt, 1}
	return &GameManager{
		board:     newBoard(boardSize),
		tray:      newBoard(traySize),
		boardSize: boardSize,
		tileCnt:   tileCnt,
		traySize:  traySize,
		tileSize:  sizeV{25, 25},
	}
}

type TileLoc struct {
	value   string
	onBoard bool
	onTray  bool
	moving  bool
	loc     gridLoc
}

func newTileLoc(v string) *TileLoc {
	return &TileLoc{
		value:   v,
		onBoard: false,
		onTray:  false,
		moving:  false,
		loc:     gridLoc{-1, -1},
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
		manager = newGameManager(sizeV{16, 16}, 16)
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
