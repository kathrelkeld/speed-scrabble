package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"unicode"

	"github.com/kathrelkeld/speed-scrabble/msg"
)

var mgr *GameManager // initiated in page setup.

// These listeners will be added/removed as needed.
var listenerMouseUp js.Func
var listenerMouseMove js.Func

type sizeV Vec
type gridLoc Vec
type canvasLoc Vec

func inTarget(loc, start, end canvasLoc) bool {
	return loc.X > start.X && loc.X < end.X && loc.Y > start.Y && loc.Y < end.Y
}

type Grid struct {
	Grid [][]*Tile
	loc  canvasLoc
	end  canvasLoc
	size sizeV
}

func newGridInner(size sizeV) [][]*Tile {
	var b [][]*Tile
	for i := 0; i < size.Y; i++ {
		b = append(b, []*Tile{})
		for j := 0; j < size.X; j++ {
			b[i] = append(b[i], nil)
		}
	}
	return b
}

func (g *Grid) get(l gridLoc) *Tile {
	return g.Grid[l.Y][l.X]
}

func (g *Grid) set(l gridLoc, tile *Tile) {
	g.Grid[l.Y][l.X] = tile
}

func (g *Grid) inBounds(l canvasLoc) bool {
	return inTarget(l, g.loc, g.end)
}

func (g *Grid) onGrid(l gridLoc) bool {
	return l.X >= 0 && l.X < g.size.X && l.Y >= 0 && l.Y < g.size.Y
}

func (g *Grid) coords(l canvasLoc) gridLoc {
	return gridLoc{
		(l.X - g.loc.X) / mgr.tileSize.X,
		(l.Y - g.loc.Y) / mgr.tileSize.Y,
	}
}

func (g *Grid) canvasStart(l gridLoc) canvasLoc {
	return cAdd(g.loc, canvasLoc(sMult(sizeV(l), mgr.tileSize)))
}

type GameManager struct {
	board      *Grid
	tray       *Grid
	tiles      []*Tile
	tileCnt    int
	tileSize   sizeV
	movingTile *Tile
	highlight  *gridLoc
	wordDir    gridLoc
}

func newGameManager(boardSize sizeV, tileCnt int) *GameManager {
	traySize := sizeV{tileCnt, 1}
	tileSize := sizeV{35, 35}
	boardStart := canvasLoc{10, 10}
	// TODO calculate where this needs to be based on size of board
	trayStart := canvasLoc{10, 600}
	return &GameManager{
		board: &Grid{
			Grid: newGridInner(boardSize),
			loc:  boardStart,
			end:  cAdd(boardStart, canvasLoc(sMult(tileSize, boardSize))),
			size: boardSize,
		},
		tray: &Grid{
			Grid: newGridInner(traySize),
			loc:  trayStart,
			end:  cAdd(trayStart, canvasLoc(sMult(tileSize, traySize))),
			size: traySize,
		},
		tileCnt:  tileCnt,
		tileSize: tileSize,
		wordDir:  gridLoc{1, 0},
	}
}

// Tile represents a single tile.  Marshalling must match the slimmer version of Tile
// from the server-side of the game.
type Tile struct {
	// Value is the letter of this tile.
	Value string
	// Points are the number of points this tile represents.
	Points int
	// Zone indicates where this tile is, e.g. on the board or moving.
	Zone int `json:"-"`
	// Idx is where this tile is on a Grid, assuming it is on a Grid.
	Idx gridLoc `json:"-"`
	// Loc is where this tile should be drawn on the canvas; may or may not match Idx.
	Loc canvasLoc `json:"-"`
	// MoveOffSet is (for ZoneMoving only) how far from the mouse cursor the tile is drawn.
	MoveOffset canvasLoc `json:"-"`
	// Invalid is whether this tile should be drawn as an invalid tile.
	Invalid bool `json:"-"`
}

const (
	ZoneNone   int = iota // initial untracked tile
	ZoneTray              // on the tray
	ZoneBoard             // on the board
	ZoneMoving            // actively moving
)

// addToBoard puts the tile onto the board at the given indices.
func (t *Tile) addToBoard(idx gridLoc) {
	if prev := mgr.board.get(idx); prev != nil {
		prev.sendToTray()
	}
	mgr.board.set(idx, t)
	t.Zone = ZoneBoard
	t.Idx = idx
	t.Loc = cAdd(mgr.board.loc, canvasLoc(sMult(mgr.tileSize, sizeV(idx))))
}

// addToTray puts the tile onto the tray at the given indices.
func (t *Tile) addToTray(idx gridLoc) {
	mgr.tray.set(idx, t)
	t.Zone = ZoneTray
	t.Idx = idx
	t.Loc = cAdd(mgr.tray.loc, canvasLoc(sMult(mgr.tileSize, sizeV(idx))))
	t.Loc = canvasLoc{
		mgr.tray.loc.X + mgr.tileSize.X*idx.X,
		mgr.tray.loc.Y + mgr.tileSize.Y*idx.Y,
	}
}

// sendToTray puts the tile onto the tray at the first available location.
func (t *Tile) sendToTray() {
	for j := 0; j < len(mgr.tray.Grid); j++ {
		for i := 0; i < len(mgr.tray.Grid[0]); i++ {
			idx := gridLoc{i, j}
			if mgr.tray.get(idx) == nil {
				t.addToTray(idx)
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
		t := mgr.board.get(c)
		if t == nil {
			fmt.Println("Verify marked a non-tile as invalid?")
			return
		}
		t.Invalid = true
	}
}

// markAllTilesValid undoes markInvalidTiles.
func markAllTilesValid() {
	for _, t := range mgr.tiles {
		t.Invalid = false
	}
}

// onTile returns a tile or nil, depending on whether there is a tile at the given location.
func onTile(l canvasLoc) *Tile {
	for _, t := range mgr.tiles {
		if l.X > t.Loc.X && l.X < t.Loc.X+mgr.tileSize.X &&
			l.Y > t.Loc.Y && l.Y < t.Loc.Y+mgr.tileSize.Y {
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
}

func listenerMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		// TODO do stuff here
		l := canvasLoc{x, y}
		if t := onTile(l); t != nil {
			clickOnTile(t, l)
		} else if mgr.board.inBounds(l) {
			highlightCanvas(l)
		}
		return nil
	})
}

func listenerKeyDown() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if mgr.highlight == nil {
			return nil
		}
		event := args[0]
		k := event.Get("key").String()
		switch k {
		case "ArrowUp":
			moveHighlight(gridLoc{0, -1})
		case "ArrowDown":
			moveHighlight(gridLoc{0, 1})
		case "ArrowLeft":
			moveHighlight(gridLoc{-1, 0})
		case "ArrowRight":
			moveHighlight(gridLoc{1, 0})
		case " ":
			toggleWordDir()
		case "Shift":
			toggleWordDir()
		case "Enter":
		case "Backspace":
			backspaceHighlight()
		case "Delete":
			backspaceHighlight()
		default:
			if len(k) == 1 && unicode.IsLetter(rune(k[0])) {
				findForHighlight(string(unicode.ToUpper(rune(k[0]))))
			}
		}
		event.Call("preventDefault")
		return nil
	})
}

// clickOnTile is called when the player clicks on a tile.
func clickOnTile(t *Tile, l canvasLoc) {
	if mgr.movingTile != nil {
		mgr.movingTile.sendToTray()
	}
	mgr.movingTile = t
	if t.Zone == ZoneBoard {
		mgr.board.set(t.Idx, nil)
	} else if t.Zone == ZoneTray {
		mgr.tray.set(t.Idx, nil)
	}
	t.Zone = ZoneMoving
	t.MoveOffset = canvasLoc{l.X - t.Loc.X, l.Y - t.Loc.Y}

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
	t.Loc = canvasLoc{l.X - t.MoveOffset.X, l.Y - t.MoveOffset.Y}
	draw()
}

// releaseTile is called when the player releases a tile.
func releaseTile(l canvasLoc) {
	if mgr.movingTile == nil {
		fmt.Println("error - release tile called without a moving tile")
		return
	}
	unhighlight()
	t := mgr.movingTile
	mgr.movingTile = nil

	if mgr.board.inBounds(l) {
		// Release tile onto board.
		coords := mgr.board.coords(l)
		t.addToBoard(coords)
		highlightCoords(coords)
	} else if mgr.tray.inBounds(l) {
		// Release tile onto tray.
		coords := mgr.tray.coords(l)
		if mgr.board.get(t.Idx) != nil {
			t.sendToTray()
		} else {
			t.addToTray(coords)
		}
	} else {
		// Return tile to tray.
		t.sendToTray()
	}

	canvas.Call("removeEventListener", "mousemove", listenerMouseMove)
	canvas.Call("removeEventListener", "mouseup", listenerMouseUp)

	markAllTilesValid()

	draw()
}

// moveHighlight moves the highlight in the given direction.
// There is definitely a highlight before this gets called.
func moveHighlight(d gridLoc) {
	// TODO: figure out whether to skip occupied squares
	newSpace := gAdd(*mgr.highlight, d)
	if mgr.board.onGrid(newSpace) {
		mgr.highlight = &newSpace
	}
	draw()
}

// findForHighlight finds a tile of value v in the tray and moves it to the highlight.
// There is definitely a highlight before this gets called.
// Does nothing if there is already a matching tile present.
func findForHighlight(v string) {
	if prev := mgr.board.get(*mgr.highlight); prev != nil && prev.Value == v {
		return
	}
	for _, t := range mgr.tiles {
		if t.Zone == ZoneTray && t.Value == v {
			t.addToBoard(*mgr.highlight)
			moveHighlight(mgr.wordDir)
			draw()
			return
		}
	}
}

func toggleWordDir() {
	mgr.wordDir = gridLoc{mgr.wordDir.Y, mgr.wordDir.X}
	draw()
}

func backspaceHighlight() {
	t := mgr.board.get(*mgr.highlight)
	t.sendToTray()
	draw()
}

// HighlightCoords highlights the square at location l, which is definitely on the board.
func highlightCoords(l gridLoc) {
	mgr.highlight = &l
	draw()
}

// HighlightCanvas highlights the square at location l, which is definitely on the board.
func highlightCanvas(l canvasLoc) {
	newH := mgr.board.coords(l)
	highlightCoords(newH)
}

func unhighlight() {
	mgr.highlight = nil
	draw()
}

// sendAllTilestoTray sends all tiles from the board into the tray.
func sendAllTilesToTray() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		for _, t := range mgr.tiles {
			if t.Zone == ZoneBoard {
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
