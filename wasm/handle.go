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

func inTarget(loc, start, end Vec) bool {
	return loc.X > start.X && loc.X < end.X && loc.Y > start.Y && loc.Y < end.Y
}

type Grid struct {
	// Grid is the slice of slices representation of the entire grid.
	Grid [][]*Tile
	// Loc is the canvas coordinates where this grid is drawn.
	Loc Vec
}

// newInnerGrid creates a slice of slices needed for a Grid struct of the given size.
func newInnerGrid(size Vec) [][]*Tile {
	var b [][]*Tile
	for i := 0; i < size.Y; i++ {
		b = append(b, []*Tile{})
		for j := 0; j < size.X; j++ {
			b[i] = append(b[i], nil)
		}
	}
	return b
}

func (g *Grid) IdxSize() Vec {
	return Vec{len(g.Grid[0]), len(g.Grid)}
}

func (g *Grid) CanvasEnd() Vec {
	size := g.IdxSize()
	return Add(g.Loc, Mult(mgr.tileSize, size))
}

func (g *Grid) Get(idx Vec) *Tile {
	return g.Grid[idx.Y][idx.X]
}

func (g *Grid) Set(idx Vec, tile *Tile) {
	g.Grid[idx.Y][idx.X] = tile
}

func (g *Grid) InCanvas(l Vec) bool {
	end := g.CanvasEnd()
	return inTarget(l, g.Loc, end)
}

func (g *Grid) InCoords(idx Vec) bool {
	size := g.IdxSize()
	return idx.X >= 0 && idx.X < size.X && idx.Y >= 0 && idx.Y < size.Y
}

func (g *Grid) coords(loc Vec) Vec {
	return Vec{
		(loc.X - g.Loc.X) / mgr.tileSize.X,
		(loc.Y - g.Loc.Y) / mgr.tileSize.Y,
	}
}

func (g *Grid) canvasStart(idx Vec) Vec {
	return Add(g.Loc, Mult(idx, mgr.tileSize))
}

// GameManager contains the local game state for the currently running game.
type GameManager struct {
	board      *Grid
	tray       *Grid
	tiles      []*Tile // All given tiles, regardless of their location.
	tileSize   Vec     // The canvas size of a single tile.
	movingTile *Tile   // Currently moving tile (or nil if none).
	highlight  *Vec    // Current board highlight index (or nil if none).
	wordDir    Vec     // Auto-advance direction (i.e. left or down).
}

// resetGameManager resets the global variable mgr with a new state for a new game.
func resetGameManager(boardSize Vec, tileCnt int) {
	traySize := Vec{tileCnt, 1} // Initial tray is one row.

	// TODO calculate where these need to be based on size of board
	tileSize := Vec{35, 35}
	boardStart := Vec{10, 10}
	trayStart := Vec{10, 600}

	mgr = &GameManager{
		board: &Grid{
			Grid: newInnerGrid(boardSize),
			Loc:  boardStart,
		},
		tray: &Grid{
			Grid: newInnerGrid(traySize),
			Loc:  trayStart,
		},
		tileSize: tileSize,
		wordDir:  Vec{1, 0},
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
	Idx Vec `json:"-"`
	// Loc is where this tile should be drawn on the canvas; may or may not match Idx.
	Loc Vec `json:"-"`
	// MoveOffSet is (for ZoneMoving only) how far from the mouse cursor the tile is drawn.
	MoveOffset Vec `json:"-"`
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
func (t *Tile) addToBoard(idx Vec) {
	if prev := mgr.board.Get(idx); prev != nil {
		prev.sendToTray()
	}
	mgr.board.Set(idx, t)
	t.Zone = ZoneBoard
	t.Idx = idx
	t.Loc = Add(mgr.board.Loc, Mult(mgr.tileSize, idx))
}

// addToTray puts the tile onto the tray at the given indices.
func (t *Tile) addToTray(idx Vec) {
	mgr.tray.Set(idx, t)
	t.Zone = ZoneTray
	t.Idx = idx
	t.Loc = Add(mgr.tray.Loc, Mult(mgr.tileSize, idx))
	t.Loc = Vec{
		mgr.tray.Loc.X + mgr.tileSize.X*idx.X,
		mgr.tray.Loc.Y + mgr.tileSize.Y*idx.Y,
	}
}

// sendToTray puts the tile onto the tray at the first available location.
func (t *Tile) sendToTray() {
	traySize := mgr.tray.IdxSize()
	for j := 0; j < traySize.Y; j++ {
		for i := 0; i < traySize.X; i++ {
			idx := Vec{i, j}
			if mgr.tray.Get(idx) == nil {
				t.addToTray(idx)
				return
			}
		}
	}
	// TODO expand downward if needed
	mgr.tray.Grid[0] = append(mgr.tray.Grid[0], t)
	t.addToTray(Vec{len(mgr.tray.Grid[0]) - 1, 0})
}

// markInvalidTiles marks the given locations on the board as invalid.
func markInvalidTiles(coords []Vec) {
	for _, c := range coords {
		t := mgr.board.Get(c)
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
func onTile(l Vec) *Tile {
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
		releaseTile(Vec{x, y})
		return nil
	})
	listenerMouseMove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		moveTile(Vec{x, y})
		return nil
	})
}

func listenerMouseDown() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		// TODO do stuff here
		l := Vec{x, y}
		if t := onTile(l); t != nil {
			clickOnTile(t, l)
		} else if mgr.board.InCanvas(l) {
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
			moveHighlight(Vec{0, -1})
		case "ArrowDown":
			moveHighlight(Vec{0, 1})
		case "ArrowLeft":
			moveHighlight(Vec{-1, 0})
		case "ArrowRight":
			moveHighlight(Vec{1, 0})
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
func clickOnTile(t *Tile, l Vec) {
	if mgr.movingTile != nil {
		mgr.movingTile.sendToTray()
	}
	mgr.movingTile = t
	if t.Zone == ZoneBoard {
		mgr.board.Set(t.Idx, nil)
	} else if t.Zone == ZoneTray {
		mgr.tray.Set(t.Idx, nil)
	}
	t.Zone = ZoneMoving
	t.MoveOffset = Vec{l.X - t.Loc.X, l.Y - t.Loc.Y}

	canvas.Call("addEventListener", "mousemove", listenerMouseMove)
	canvas.Call("addEventListener", "mouseup", listenerMouseUp)

	markAllTilesValid()
	draw()
}

// moveTile is called when the player moves a tile.
func moveTile(l Vec) {
	if mgr.movingTile == nil {
		fmt.Println("error - move tile called without a moving tile")
		return
	}
	t := mgr.movingTile
	t.Loc = Vec{l.X - t.MoveOffset.X, l.Y - t.MoveOffset.Y}
	draw()
}

// releaseTile is called when the player releases a tile.
func releaseTile(l Vec) {
	if mgr.movingTile == nil {
		fmt.Println("error - release tile called without a moving tile")
		return
	}
	unhighlight()
	t := mgr.movingTile
	mgr.movingTile = nil

	if mgr.board.InCanvas(l) {
		// Release tile onto board.
		coords := mgr.board.coords(l)
		t.addToBoard(coords)
		highlightCoords(coords)
	} else if mgr.tray.InCanvas(l) {
		// Release tile onto tray.
		coords := mgr.tray.coords(l)
		if mgr.board.Get(t.Idx) != nil {
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
func moveHighlight(d Vec) {
	// TODO: figure out whether to skip occupied squares
	newSpace := Add(*mgr.highlight, d)
	if mgr.board.InCoords(newSpace) {
		mgr.highlight = &newSpace
	}
	draw()
}

// findForHighlight finds a tile of value v in the tray and moves it to the highlight.
// There is definitely a highlight before this gets called.
// Does nothing if there is already a matching tile present.
func findForHighlight(v string) {
	if prev := mgr.board.Get(*mgr.highlight); prev != nil && prev.Value == v {
		moveHighlight(mgr.wordDir)
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
	mgr.wordDir = Vec{mgr.wordDir.Y, mgr.wordDir.X}
	draw()
}

func backspaceHighlight() {
	if t := mgr.board.Get(*mgr.highlight); t != nil {
		t.sendToTray()
	}
	moveHighlight(ScaleUp(mgr.wordDir, -1))
	draw()
}

// HighlightCoords highlights the square at location l, which is definitely on the board.
func highlightCoords(l Vec) {
	mgr.highlight = &l
	draw()
}

// HighlightCanvas highlights the square at location l, which is definitely on the board.
func highlightCanvas(l Vec) {
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
