package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

// GameManager contains the local game state for the currently running game.
type GameManager struct {
	board     *Grid
	tray      *Grid
	tiles     []*Tile    // All given tiles, regardless of their location.
	tileSize  Vec        // The canvas size of a single tile.
	move      *Move      // Current move action.
	highlight *Highlight // Current board highlight.
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
			Zone: ZoneBoard,
		},
		tray: &Grid{
			Grid: newInnerGrid(traySize),
			Loc:  trayStart,
			Zone: ZoneTray,
		},
		move: &Move{},
		highlight: &Highlight{
			dir: Vec{1, 0},
		},
		tileSize: tileSize,
	}
}

// ShiftBoard moves the entire contents of the board in the given direction, sending the
// tiles back to the tray if they fall off.
func ShiftBoard(d Vec) {
	type Work struct {
		t *Tile
		l Vec
	}
	var toMove []Work
	for _, t := range mgr.tiles {
		if t.Zone == ZoneBoard {
			next := Add(t.Idx, d)
			if !mgr.board.InCoords(next) {
				return
			} else {
				toMove = append(toMove, Work{t, next})
			}
		}
	}
	for _, w := range toMove {
		w.t.addToBoard(w.l)
	}
	draw()
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

// pickUp removes the tile from its current grid, if any.
func (t *Tile) pickUp() {
	switch t.Zone {
	case ZoneTray:
		mgr.tray.Set(t.Idx, nil)
	case ZoneBoard:
		mgr.board.Set(t.Idx, nil)
	}
	t.Zone = ZoneNone
}

// addToBoard puts the tile onto the board at the given indices.
func (t *Tile) addToBoard(idx Vec) {
	if prev := mgr.board.Get(idx); prev != nil {
		prev.sendToTray()
	}
	mgr.board.AddTile(t, idx)
}

// addToTray puts the tile onto the tray at the given indices.
func (t *Tile) addToTray(idx Vec) {
	if mgr.tray.Get(idx) != nil {
		t.sendToTray()
	}
	mgr.tray.AddTile(t, idx)
}

// sendToTray puts the tile onto the tray at the first available location.
func (t *Tile) sendToTray() {
	t.pickUp()
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
	mgr.tray.Grid[0] = append(mgr.tray.Grid[0], nil)
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

type Move struct {
	active     bool
	onTile     bool
	tile       *Tile
	startZone  int // Where the tile was before it moved.
	startIdx   Vec // Coordinates of the tile before it moved.
	startClick Vec // Where on the canvas was clicked to start the move.
	hasMoved   bool
}

type Highlight struct {
	active bool
	idx    Vec // Coordinates on the board where the highlight starts.
	dir    Vec // Direction to advance when typing.
}

// initializeListners sets the global listener variables for input.
func initializeListeners() {
	listenerMouseUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !mgr.move.active {
			fmt.Println("saw a mouseUp with no active move recorded!")
			return nil
		}
		event := args[0]
		x := event.Get("offsetX").Int()
		y := event.Get("offsetY").Int()
		l := Vec{x, y}
		if mgr.move.onTile {
			releaseTile(l)
		}
		releaseUpdateHighlight(l)
		mgr.move = &Move{}
		canvas.Call("removeEventListener", "mousemove", listenerMouseMove)
		canvas.Call("removeEventListener", "mouseup", listenerMouseUp)
		return nil
	})
	listenerMouseMove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !mgr.move.active {
			fmt.Println("saw a mouseMove with no active move recorded!")
			return nil
		}
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
		l := Vec{x, y}
		mgr.move = &Move{
			active:     true,
			startClick: l,
		}
		canvas.Call("addEventListener", "mousemove", listenerMouseMove)
		canvas.Call("addEventListener", "mouseup", listenerMouseUp)

		if t := onTile(l); t != nil {
			clickOnTile(t, l)
		}

		return nil
	})
}

func listenerKeyDown() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !mgr.highlight.active {
			return nil
		}
		event := args[0]
		k := event.Get("key").String()
		switch k {
		case "ArrowUp":
			//moveHighlight(Vec{0, -1})
			ShiftBoard(Vec{0, -1})
		case "ArrowDown":
			//moveHighlight(Vec{0, 1})
			ShiftBoard(Vec{0, 1})
		case "ArrowLeft":
			//moveHighlight(Vec{-1, 0})
			ShiftBoard(Vec{-1, 0})
		case "ArrowRight":
			//moveHighlight(Vec{1, 0})
			ShiftBoard(Vec{1, 0})
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

// clickOnTile is called when the player clicks on a tile at the given canvas location.
func clickOnTile(t *Tile, l Vec) {
	t.pickUp()
	mgr.move.onTile = true
	mgr.move.tile = t
	mgr.move.startZone = t.Zone
	mgr.move.startIdx = t.Idx
	t.Zone = ZoneMoving
	t.MoveOffset = Sub(l, t.Loc)

	draw()
}

// moveTile is called when the player moves a tile with the mouse.
func moveTile(l Vec) {
	mgr.move.hasMoved = true
	if mgr.move.onTile {
		t := mgr.move.tile
		t.Loc = Vec{l.X - t.MoveOffset.X, l.Y - t.MoveOffset.Y}
		draw()
	}
}

// releaseTile is called when the player releases a tile.
func releaseTile(l Vec) {
	t := mgr.move.tile

	switch {
	case mgr.board.InCanvas(l):
		// Release tile onto board.
		coords := mgr.board.coords(l)
		t.addToBoard(coords)
		if !mgr.highlight.active {
			highlightCoords(coords)
		}
	case mgr.tray.InCanvas(l):
		// Release tile onto tray.
		coords := mgr.tray.coords(l)
		t.addToTray(coords)
	default:
		t.sendToTray()
	}

	markAllTilesValid()
	draw()
}

// releaseUpdateHighlight is called when the player releases a click/drag.
func releaseUpdateHighlight(l Vec) {
	switch {
	case mgr.board.InCanvas(l):
		coords := mgr.board.coords(l)
		if mgr.move.startClick == l && !mgr.move.hasMoved {
			if mgr.highlight.active && mgr.highlight.idx == coords {
				// click was on highlight, so toggle instead.
				toggleWordDir()
			} else {
				highlightCoords(coords)
			}
		}
	default:
		unhighlight()
	}
	draw()
}

// moveHighlight moves the highlight in the given direction.
// There is definitely a highlight before this gets called.
func moveHighlight(d Vec) {
	// TODO: figure out whether to skip occupied squares
	newSpace := Add(mgr.highlight.idx, d)
	if mgr.board.InCoords(newSpace) {
		mgr.highlight.idx = newSpace
	}
	draw()
}

// findForHighlight finds a tile of value v in the tray and moves it to the highlight.
// There is definitely a highlight before this gets called.
// Does nothing if there is already a matching tile present.
func findForHighlight(v string) {
	if prev := mgr.board.Get(mgr.highlight.idx); prev != nil && prev.Value == v {
		moveHighlight(mgr.highlight.dir)
		return
	}
	for _, t := range mgr.tiles {
		if t.Zone == ZoneTray && t.Value == v {
			t.addToBoard(mgr.highlight.idx)
			moveHighlight(mgr.highlight.dir)
			draw()
			return
		}
	}
}

func toggleWordDir() {
	mgr.highlight.dir = Vec{mgr.highlight.dir.Y, mgr.highlight.dir.X}
	draw()
}

func backspaceHighlight() {
	if t := mgr.board.Get(mgr.highlight.idx); t != nil {
		t.sendToTray()
	}
	moveHighlight(ScaleUp(mgr.highlight.dir, -1))
	draw()
}

// HighlightCoords highlights the square at location l, which is definitely on the board.
func highlightCoords(l Vec) {
	mgr.highlight.active = true
	mgr.highlight.idx = l
	draw()
}

// HighlightCanvas highlights the square at location l, which is definitely on the board.
func highlightCanvas(l Vec) {
	newH := mgr.board.coords(l)
	highlightCoords(newH)
}

func unhighlight() {
	mgr.highlight.active = false
	mgr.highlight.idx = Vec{-1, -1}
	mgr.highlight.dir = Vec{1, 0}
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
		unhighlight()
		markAllTilesValid()
		draw()
		return nil
	})
}

// shuffleTiles reorders the tiles in the tray.
func shuffleTiles() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var ts []*Tile
		for _, t := range mgr.tiles {
			if t.Zone == ZoneTray {
				t.pickUp()
				ts = append(ts, t)
			}
		}
		rand.Shuffle(len(ts), func(i, j int) {
			ts[i], ts[j] = ts[j], ts[i]
		})
		for _, t := range ts {
			t.sendToTray()
		}
		markAllTilesValid()
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
