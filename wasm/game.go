package main

import (
	"fmt"
	"math/rand"
)

const (
	StateNoGame int = iota
	StateHasGame
	StatePlaying
	StateGameOver
)

// GameManager contains the local game state for the currently running game.
type GameManager struct {
	state     int
	board     *Grid
	tray      *Grid
	tiles     []*Tile // All given tiles, regardless of their location.
	tileSize  Vec     // The canvas size of a single tile.
	badWords  []Word
	move      *Move      // Current move action.
	highlight *Highlight // Current board highlight.
	listens   Listeners
	ctx       Context
	canvas    Canvas
}

// NewGameManager resets the global variable mgr with a new state for a new game.
func NewGameManager(boardSize Vec, tileCnt int) *GameManager {
	traySize := Vec{tileCnt, 1} // Initial tray is one row.

	// TODO calculate where these need to be based on size of board
	tileSize := Vec{35, 35}
	boardStart := Vec{10, 10}
	trayStart := Vec{10, 600}

	mgr := &GameManager{
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
	mgr.board.mgr = mgr
	mgr.tray.mgr = mgr
	return mgr
}

func (mgr *GameManager) Reset() {
	// TODO figure out what to do about tileCnt
	next := NewGameManager(mgr.board.IdxSize(), 16)
	mgr.state = next.state
	mgr.board = next.board
	mgr.tray = next.tray
	mgr.tiles = next.tiles
	mgr.badWords = next.badWords
	mgr.move = next.move
	mgr.highlight = next.highlight
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
	// State is whether this tile should be drawn as an invalid or unused tile.
	State int `json:"-"`
	// mgr references the parent GameManager which added this tile.
	mgr *GameManager
}

const (
	ZoneNone      int = iota // initial untracked tile
	ZoneTray                 // on the tray
	ZoneBoard                // on the board
	ZoneMoving               // actively moving
	ZoneOffScreen            // not on any known area
)

const (
	TileStateValid int = iota
	TileStateInvalid
	TileStateUnused
)

// pickUp removes the tile from its current grid, if any.
func (t *Tile) pickUp() {
	switch t.Zone {
	case ZoneTray:
		t.mgr.tray.Set(t.Idx, nil)
	case ZoneBoard:
		t.mgr.board.Set(t.Idx, nil)
	}
	t.Zone = ZoneNone
}

// addToBoard puts the tile onto the board at the given indices.
func (t *Tile) addToBoard(idx Vec) {
	if prev := t.mgr.board.Get(idx); prev != nil {
		prev.sendToTray()
	}
	t.mgr.board.AddTile(t, idx)
}

// addToTray puts the tile onto the tray at the given indices.
func (t *Tile) addToTray(idx Vec) {
	if t.mgr.tray.Get(idx) != nil {
		t.sendToTray()
	}
	t.mgr.tray.AddTile(t, idx)
}

// sendToTray puts the tile onto the tray at the first available location.
func (t *Tile) sendToTray() {
	t.pickUp()
	traySize := t.mgr.tray.IdxSize()
	for j := 0; j < traySize.Y; j++ {
		for i := 0; i < traySize.X; i++ {
			idx := Vec{i, j}
			if t.mgr.tray.Get(idx) == nil {
				t.addToTray(idx)
				return
			}
		}
	}
	// TODO expand downward if needed
	t.mgr.tray.Grid[0] = append(t.mgr.tray.Grid[0], nil)
	t.addToTray(Vec{len(t.mgr.tray.Grid[0]) - 1, 0})
}

func (mgr *GameManager) markTiles(state int, coords []Vec) {
	for _, c := range coords {
		t := mgr.board.Get(c)
		if t == nil {
			fmt.Println("Verify marked a non-tile as invalid?")
			return
		}
		t.State = state
	}
}

// markInvalidTiles marks the given locations on the board as invalid.
func (mgr *GameManager) markInvalidAndUnusedTiles(invalid, unused []Vec, badWords []Word) {
	mgr.markTiles(TileStateInvalid, invalid)
	mgr.markTiles(TileStateUnused, unused)
	mgr.badWords = badWords
}

// markAllTilesValid undoes markInvalidTiles.
func (mgr *GameManager) unmarkAllTiles() {
	mgr.badWords = []Word{}
	for _, t := range mgr.tiles {
		t.State = TileStateValid
	}
}

// onTile returns a tile or nil, depending on whether there is a tile at the given location.
func (mgr *GameManager) onTile(l Vec) *Tile {
	for _, t := range mgr.tiles {
		if l.X > t.Loc.X && l.X < t.Loc.X+mgr.tileSize.X &&
			l.Y > t.Loc.Y && l.Y < t.Loc.Y+mgr.tileSize.Y {
			return t
		}
	}
	return nil
}

// ShiftBoard moves the entire contents of the board in the given direction, sending the
// tiles back to the tray if they fall off.
func (mgr *GameManager) ShiftBoard(d Vec) {
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
}

// sendAllTilestoTray sends all tiles from the board into the tray.
func (mgr *GameManager) sendAllTilesToTray() {
	for _, t := range mgr.tiles {
		if t.Zone == ZoneBoard {
			t.sendToTray()
		}
	}
	mgr.unhighlight()
	mgr.unmarkAllTiles()
}

// shuffleTiles reorders the tiles in the tray.
func (mgr *GameManager) shuffleTiles() {
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
}
