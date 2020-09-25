package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var manager *GameManager

type Vec struct {
	X int
	Y int
}

type Tile struct {
	Value  string
	Points int
}

type sizeV Vec
type gridLoc Vec
type canvasLoc Vec
type board [][]*TileLoc

func newBoard(size sizeV) board {
	var b board
	for i := 0; i < size.Y; i++ {
		b = append(b, []*TileLoc{})
		for j := 0; j < size.X; j++ {
			b[i] = append(b[i], nil)
		}
	}
	return b
}

type GameManager struct {
	board     board
	boardLoc  canvasLoc
	tray      board
	trayLoc   canvasLoc
	tiles     []*TileLoc
	boardSize sizeV
	traySize  sizeV
	tileCnt   int
	tileSize  sizeV
}

func newGameManager(boardSize sizeV, tileCnt int) *GameManager {
	traySize := sizeV{tileCnt, 1}
	return &GameManager{
		board:    newBoard(boardSize),
		boardLoc: canvasLoc{10, 10},
		tray:     newBoard(traySize),
		// TODO calculate where this needs to be based on size of board
		trayLoc:   canvasLoc{10, 450},
		boardSize: boardSize,
		tileCnt:   tileCnt,
		traySize:  traySize,
		tileSize:  sizeV{25, 25},
	}
}

type TileLoc struct {
	value     string
	region    int
	gridLoc   gridLoc
	canvasLoc canvasLoc
}

const (
	OnNone int = iota
	OnTray
	OnBoard
	OnMoving
)

func newTileLoc(v string) *TileLoc {
	return &TileLoc{
		value:     v,
		region:    OnNone,
		gridLoc:   gridLoc{-1, -1},
		canvasLoc: canvasLoc{-1, -1},
	}
}

func onTile(l canvasLoc) *TileLoc {
	for _, t := range manager.tiles {
		if l.X > t.canvasLoc.X && l.X < t.canvasLoc.X+manager.tileSize.X &&
			l.Y > t.canvasLoc.Y && l.Y < t.canvasLoc.Y+manager.tileSize.Y {
			return t
		}
	}
	return nil
}

func coordsOnTray(gl gridLoc) canvasLoc {
	return canvasLoc{
		manager.trayLoc.X + manager.tileSize.X*gl.X,
		manager.trayLoc.Y + manager.tileSize.Y*gl.Y,
	}
}

func coordsOnBoard(gl gridLoc) canvasLoc {
	return canvasLoc{
		manager.boardLoc.X + manager.tileSize.X*gl.X,
		manager.boardLoc.Y + manager.tileSize.Y*gl.Y,
	}
}

func sendTileToTray(t *TileLoc) {
	t.region = OnTray
	for j := 0; j < len(manager.tray); j++ {
		for i := 0; i < len(manager.tray[0]); i++ {
			if manager.tray[j][i] == nil {
				manager.tray[j][i] = t
				t.gridLoc = gridLoc{i, j}
				t.canvasLoc = coordsOnTray(t.gridLoc)
				return
			}
		}
	}
	// TODO expand downward if needed
	manager.tray[0] = append(manager.tray[0], t)
	t.gridLoc = gridLoc{len(manager.tray[0]) - 1, 0}
}

func sendTilesToTray(ts []*TileLoc) {
	for _, t := range ts {
		sendTileToTray(t)
	}
}

func sendAllTilesToTray() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for _, t := range manager.tiles {
			if t.region == OnBoard {
				sendTileToTray(t)
			}
		}
		return nil
	})
}

func requestNewTile() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		websocketSendEmpty(msg.AddTile)
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
		m, _ := msg.NewSocketData(msg.Verify, nil)
		websocketSend(m)
		return nil
	})
}

func handleSocketMsg(t msg.Type, data []byte) int {
	switch t {
	case msg.OK:
		if len(onOk) > 0 {
			onOk[0].Invoke()
			onOk = onOk[1:]
		}
	case msg.Error:
	case msg.Start:
		if manager != nil {
			// TODO delete old manager
		}
		// TODO tie to actual game size
		manager = newGameManager(sizeV{16, 16}, 16)
		var tiles []*Tile
		err := json.Unmarshal(data, &tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", tiles)
		for _, tile := range tiles {
			manager.tiles = append(manager.tiles, newTileLoc(tile.Value))
		}
		sendTilesToTray(manager.tiles)
		draw()
	case msg.AddTile:
		var tile Tile
		err := json.Unmarshal(data, &t)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		manager.tiles = append(manager.tiles, newTileLoc(tile.Value))
		fmt.Println("Adding new tile:", tile.Value)
	case msg.Score:

	case msg.GameStatus:
		var s msg.GameInfo
		err := json.Unmarshal(data, &s)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("Game:", s.GameName)
	}
	return 0
}
