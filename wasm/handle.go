package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var manager *GameManager

var listenerMouseDown js.Func
var listenerMouseUp js.Func
var listenerMouseMove js.Func

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

func isInTarget(loc, start, end canvasLoc) bool {
	return loc.X > start.X && loc.X < end.X && loc.Y > start.Y && loc.Y < end.Y
}

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
	board      board
	boardLoc   canvasLoc
	boardEnd   canvasLoc
	tray       board
	trayLoc    canvasLoc
	trayEnd    canvasLoc
	tiles      []*TileLoc
	boardSize  sizeV
	traySize   sizeV
	tileCnt    int
	tileSize   sizeV
	movingTile *TileLoc
}

func newGameManager(boardSize sizeV, tileCnt int) *GameManager {
	traySize := sizeV{tileCnt, 1}
	tileSize := sizeV{25, 25}
	return &GameManager{
		board:    newBoard(boardSize),
		boardLoc: canvasLoc{10, 10},
		boardEnd: canvasLoc{10 + tileSize.X*boardSize.X, 10 + tileSize.Y*boardSize.Y},
		tray:     newBoard(traySize),
		// TODO calculate where this needs to be based on size of board
		trayLoc:   canvasLoc{10, 450},
		trayEnd:   canvasLoc{10 + tileSize.X*tileCnt, 10 + tileSize.Y},
		boardSize: boardSize,
		tileCnt:   tileCnt,
		traySize:  traySize,
		tileSize:  tileSize,
	}
}

type TileLoc struct {
	value      string
	region     int
	gridLoc    gridLoc
	canvasLoc  canvasLoc
	moveOffset canvasLoc
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
	manager.traySize.X += 1
	t.gridLoc = gridLoc{len(manager.tray[0]) - 1, 0}
	t.canvasLoc = coordsOnTray(t.gridLoc)
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

func initializeListeners() {
	listenerMouseUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		releaseTile(canvasLoc{x, y})
		return nil
	})
	listenerMouseMove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		moveTile(canvasLoc{x, y})
		return nil
	})
	listenerMouseDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		// TODO do stuff here
		l := canvasLoc{x, y}
		if t := onTile(l); t != nil {
			clickOnTile(t, l)
		}
		return nil
	})
}

func clickOnTile(t *TileLoc, l canvasLoc) {
	if manager.movingTile != nil {
		sendTileToTray(manager.movingTile)
	}
	manager.movingTile = t
	if t.region == OnTray {
		manager.tray[t.gridLoc.Y][t.gridLoc.X] = nil
	} else if t.region == OnBoard {
		manager.board[t.gridLoc.Y][t.gridLoc.X] = nil
	}
	t.region = OnMoving
	t.moveOffset = canvasLoc{l.X - t.canvasLoc.X, l.Y - t.canvasLoc.Y}

	canvas.Call("addEventListener", "mousemove", listenerMouseMove)
	canvas.Call("addEventListener", "mouseup", listenerMouseUp)

}

func moveTile(l canvasLoc) {
	if manager.movingTile == nil {
		fmt.Println("error - move tile called without a moving tile")
		return
	}
	t := manager.movingTile
	t.canvasLoc = canvasLoc{l.X - t.moveOffset.X, l.Y - t.moveOffset.Y}
	draw()
}

func releaseTile(l canvasLoc) {
	if manager.movingTile == nil {
		fmt.Println("error - release tile called without a moving tile")
		return
	}
	t := manager.movingTile
	manager.movingTile = nil

	if isInTarget(l, manager.boardLoc, manager.boardEnd) {
		// Release tile onto board.
		boardCoords := gridLoc{
			(l.X - manager.boardLoc.X) / manager.tileSize.X,
			(l.Y - manager.boardLoc.Y) / manager.tileSize.Y,
		}
		if manager.board[t.gridLoc.Y][t.gridLoc.X] != nil {
			// TODO swap the tiles?
			sendTileToTray(t)
		}
		manager.board[t.gridLoc.Y][t.gridLoc.X] = t
		t.gridLoc = boardCoords
		t.canvasLoc = coordsOnBoard(boardCoords)
		t.region = OnBoard
	} else if isInTarget(l, manager.trayLoc, manager.trayEnd) {
		// Release tile onto tray.
		trayCoords := gridLoc{
			(l.X - manager.trayLoc.X) / manager.tileSize.X,
			(l.Y - manager.trayLoc.Y) / manager.tileSize.Y,
		}
		if manager.board[t.gridLoc.Y][t.gridLoc.X] != nil {
			sendTileToTray(t)
		}
		manager.board[t.gridLoc.Y][t.gridLoc.X] = t
		t.gridLoc = trayCoords
		t.canvasLoc = coordsOnTray(trayCoords)
		t.region = OnTray
	} else {
		// Return untracked tile to tray.
		sendTileToTray(t)
	}

	canvas.Call("removeEventListener", "mousemove", listenerMouseMove)
	canvas.Call("removeEventListener", "mouseup", listenerMouseUp)

	draw()
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
			loc := newTileLoc(tile.Value)
			manager.tiles = append(manager.tiles, loc)
			sendTileToTray(loc)
		}
		draw()
	case msg.AddTile:
		var tile Tile
		err := json.Unmarshal(data, &tile)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		loc := newTileLoc(tile.Value)
		manager.tiles = append(manager.tiles, loc)
		sendTileToTray(loc)
		draw()
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
