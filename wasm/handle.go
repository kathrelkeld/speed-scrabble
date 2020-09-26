package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var mgr *GameManager // initiated in page setup.

var listenerMouseDown js.Func
var listenerMouseUp js.Func
var listenerMouseMove js.Func

type Tile struct {
	Value  string
	Points int
}

type sizeV Vec
type gridLoc Vec
type canvasLoc Vec

type Grid struct {
	Grid [][]*TileLoc
	loc  canvasLoc
	end  canvasLoc
	size sizeV
}

func inTarget(loc, start, end canvasLoc) bool {
	return loc.X > start.X && loc.X < end.X && loc.Y > start.Y && loc.Y < end.Y
}

func (g *Grid) inBounds(loc canvasLoc) bool {
	return inTarget(loc, g.loc, g.end)
}

func newGridInner(size sizeV) [][]*TileLoc {
	var b [][]*TileLoc
	for i := 0; i < size.Y; i++ {
		b = append(b, []*TileLoc{})
		for j := 0; j < size.X; j++ {
			b[i] = append(b[i], nil)
		}
	}
	return b
}

type GameManager struct {
	board      Grid
	tray       Grid
	tiles      []*TileLoc
	tileCnt    int
	tileSize   sizeV
	movingTile *TileLoc
	highlight  *gridLoc
}

func newGameManager(boardSize sizeV, tileCnt int) *GameManager {
	traySize := sizeV{tileCnt, 1}
	tileSize := sizeV{35, 35}
	boardStart := canvasLoc{10, 10}
	// TODO calculate where this needs to be based on size of board
	trayStart := canvasLoc{10, 600}
	return &GameManager{
		board: Grid{
			Grid: newGridInner(boardSize),
			loc:  boardStart,
			end:  cAdd(boardStart, canvasLoc(sMult(tileSize, boardSize))),
			size: boardSize,
		},
		tray: Grid{
			Grid: newGridInner(traySize),
			loc:  trayStart,
			end:  cAdd(trayStart, canvasLoc(sMult(tileSize, traySize))),
			size: traySize,
		},
		tileCnt:  tileCnt,
		tileSize: tileSize,
	}
}

type TileLoc struct {
	Value      string
	region     int
	gridLoc    gridLoc
	canvasLoc  canvasLoc
	moveOffset canvasLoc
	invalid    bool
}

const (
	OnNone int = iota
	OnTray
	OnBoard
	OnMoving
)

func newTileLoc(v string) *TileLoc {
	return &TileLoc{
		Value:     v,
		region:    OnNone,
		gridLoc:   gridLoc{-1, -1},
		canvasLoc: canvasLoc{-1, -1},
	}
}

// addToBoard puts the tile onto the board at the given location.
func (t *TileLoc) addToBoard(gl gridLoc) {
	mgr.board.Grid[gl.Y][gl.X] = t
	t.region = OnBoard
	t.gridLoc = gl
	t.canvasLoc = cAdd(mgr.board.loc, canvasLoc(sMult(mgr.tileSize, sizeV(gl))))
}

// addToTray puts the tile onto the tray at the given location.
func (t *TileLoc) addToTray(gl gridLoc) {
	mgr.tray.Grid[gl.Y][gl.X] = t
	t.region = OnTray
	t.gridLoc = gl
	t.canvasLoc = cAdd(mgr.tray.loc, canvasLoc(sMult(mgr.tileSize, sizeV(gl))))
	t.canvasLoc = canvasLoc{
		mgr.tray.loc.X + mgr.tileSize.X*gl.X,
		mgr.tray.loc.Y + mgr.tileSize.Y*gl.Y,
	}
}

// sendToTray puts the tile onto the tray at the first available location.
func (t *TileLoc) sendToTray() {
	for j := 0; j < len(mgr.tray.Grid); j++ {
		for i := 0; i < len(mgr.tray.Grid[0]); i++ {
			if mgr.tray.Grid[j][i] == nil {
				t.addToTray(gridLoc{i, j})
				return
			}
		}
	}
	// TODO expand downward if needed
	mgr.tray.Grid[0] = append(mgr.tray.Grid[0], t)
	mgr.tray.size.X += 1
	t.addToTray(gridLoc{len(mgr.tray.Grid[0]) - 1, 0})
}

// markInvalidTiles marks the given locations on the board as invalid.
func markInvalidTiles(coords []gridLoc) {
	for _, c := range coords {
		t := mgr.board.Grid[c.Y][c.X]
		if t == nil {
			fmt.Println("Verify marked a non-tile as invalid?")
			return
		}
		t.invalid = true
	}
}

// markAllTilesValid undoes markInvalidTiles.
func markAllTilesValid() {
	for _, t := range mgr.tiles {
		t.invalid = false
	}
}

// onTile returns a tile or nil, depending on whether there is a tile at the given location.
func onTile(l canvasLoc) *TileLoc {
	for _, t := range mgr.tiles {
		if l.X > t.canvasLoc.X && l.X < t.canvasLoc.X+mgr.tileSize.X &&
			l.Y > t.canvasLoc.Y && l.Y < t.canvasLoc.Y+mgr.tileSize.Y {
			return t
		}
	}
	return nil
}

// initializeListners sets the global listener variables for input.
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
		} else if mgr.board.inBounds(l) {
			highlightSpace(l)
		}
		return nil
	})
}

// clickOnTile is called when the player clicks on a tile.
func clickOnTile(t *TileLoc, l canvasLoc) {
	if mgr.movingTile != nil {
		mgr.movingTile.sendToTray()
	}
	mgr.movingTile = t
	if t.region == OnBoard {
		mgr.board.Grid[t.gridLoc.Y][t.gridLoc.X] = nil
	} else if t.region == OnTray {
		mgr.tray.Grid[t.gridLoc.Y][t.gridLoc.X] = nil
	}
	t.region = OnMoving
	t.moveOffset = canvasLoc{l.X - t.canvasLoc.X, l.Y - t.canvasLoc.Y}

	canvas.Call("addEventListener", "mousemove", listenerMouseMove)
	canvas.Call("addEventListener", "mouseup", listenerMouseUp)

	markAllTilesValid()
	draw()
}

// moveTile is called when the player moves a tile.
func moveTile(l canvasLoc) {
	if mgr.movingTile == nil {
		fmt.Println("error - move tile called without a moving tile")
		return
	}
	t := mgr.movingTile
	t.canvasLoc = canvasLoc{l.X - t.moveOffset.X, l.Y - t.moveOffset.Y}
	draw()
}

// releaseTile is called when the player releases a tile.
func releaseTile(l canvasLoc) {
	if mgr.movingTile == nil {
		fmt.Println("error - release tile called without a moving tile")
		return
	}
	t := mgr.movingTile
	mgr.movingTile = nil

	if mgr.board.inBounds(l) {
		// Release tile onto board.
		boardCoords := gridLoc{
			(l.X - mgr.board.loc.X) / mgr.tileSize.X,
			(l.Y - mgr.board.loc.Y) / mgr.tileSize.Y,
		}
		if mgr.board.Grid[t.gridLoc.Y][t.gridLoc.X] != nil {
			// TODO swap the tiles?
			t.sendToTray()
		} else {
			t.addToBoard(boardCoords)
		}
	} else if mgr.tray.inBounds(l) {
		// Release tile onto tray.
		trayCoords := gridLoc{
			(l.X - mgr.tray.loc.X) / mgr.tileSize.X,
			(l.Y - mgr.tray.loc.Y) / mgr.tileSize.Y,
		}
		if mgr.board.Grid[t.gridLoc.Y][t.gridLoc.X] != nil {
			t.sendToTray()
		} else {
			t.addToTray(trayCoords)
		}
	} else {
		// Return untracked tile to tray.
		t.sendToTray()
	}

	canvas.Call("removeEventListener", "mousemove", listenerMouseMove)
	canvas.Call("removeEventListener", "mouseup", listenerMouseUp)

	markAllTilesValid()

	draw()
}

func highlightSpace(l canvasLoc) {
	//mgr.highlight = l
	draw()
}

func unhighlight() {
	mgr.highlight = nil
	draw()
}

// sendAllTilestoTray sends all tiles from the board into the tray.
func sendAllTilesToTray() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for _, t := range mgr.tiles {
			if t.region == OnBoard {
				t.sendToTray()
			}
		}
		draw()
		return nil
	})
}

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
		mgr = newGameManager(sizeV{16, 16}, 16)
		var tiles []*Tile
		err := json.Unmarshal(data, &tiles)
		if err != nil {
			fmt.Println("Error reading game status:", err)
			return 1
		}
		fmt.Println("current tiles:", tiles)
		for _, tile := range tiles {
			loc := newTileLoc(tile.Value)
			mgr.tiles = append(mgr.tiles, loc)
			loc.sendToTray()
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
		mgr.tiles = append(mgr.tiles, loc)
		loc.sendToTray()
		draw()
	case msg.Score:

	case msg.Invalid:
		var invalid []gridLoc
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
