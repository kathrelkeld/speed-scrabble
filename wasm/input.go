package main

import (
	"fmt"
	"syscall/js"
	"unicode"
)

// These listeners will be added/removed as needed.
var listenerMouseUp js.Func
var listenerMouseMove js.Func

type Move struct {
	active     bool
	onTile     bool
	tile       *Tile
	startZone  int // Where the tile was before it moved.
	startIdx   Vec // Coordinates of the tile before it moved.
	startClick Vec // Where on the canvas was clicked to start the move.
	hasMoved   bool
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
			if event.Get("ctrlKey").Bool() {
				ShiftBoard(Vec{0, -1})
			} else {
				moveHighlight(Vec{0, -1})
			}
		case "ArrowDown":
			if event.Get("ctrlKey").Bool() {
				ShiftBoard(Vec{0, 1})
			} else {
				moveHighlight(Vec{0, 1})
			}
		case "ArrowLeft":
			if event.Get("ctrlKey").Bool() {
				ShiftBoard(Vec{-1, 0})
			} else {
				moveHighlight(Vec{-1, 0})
			}
		case "ArrowRight":
			if event.Get("ctrlKey").Bool() {
				ShiftBoard(Vec{1, 0})
			} else {
				moveHighlight(Vec{1, 0})
			}
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
