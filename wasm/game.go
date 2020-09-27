package main

import (
	"fmt"
	"math/rand"
	"syscall/js"
)

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
