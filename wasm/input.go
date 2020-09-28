package main

import (
	"fmt"
	"syscall/js"
	"unicode"
)

type Move struct {
	active     bool
	onTile     bool
	tile       *Tile
	startZone  int // Where the tile was before it moved.
	startIdx   Vec // Coordinates of the tile before it moved.
	startClick Vec // Where on the canvas was clicked to start the move.
	hasMoved   bool
}

type ClickInfo struct {
	zone   int
	coords Vec
}

func clickOffset(event js.Value) Vec {
	return Vec{event.Get("offsetX").Int(), event.Get("offsetY").Int()}
}

func clickInfo(l Vec) ClickInfo {
	switch {
	case mgr.board.InCanvas(l):
		return ClickInfo{ZoneBoard, mgr.board.coords(l)}
	case mgr.tray.InCanvas(l):
		return ClickInfo{ZoneTray, mgr.tray.coords(l)}
	default:
		return ClickInfo{ZoneOffScreen, Vec{-1, -1}}
	}
}

func onMouseDown(event js.Value) {
	l := clickOffset(event)
	mgr.move = &Move{
		active:     true,
		startClick: l,
	}
	mgr.listens.InMove()

	if t := onTile(l); t != nil {
		// Set this tile as moving.
		t.pickUp()
		mgr.move.onTile = true
		mgr.move.tile = t
		mgr.move.startZone = t.Zone
		mgr.move.startIdx = t.Idx
		t.Zone = ZoneMoving
		t.MoveOffset = Sub(l, t.Loc)
	}
}

func onMouseMove(event js.Value) {
	if !mgr.move.active {
		fmt.Println("saw a mouseMove with no active move recorded!")
		return
	}
	l := clickOffset(event)
	mgr.move.hasMoved = true
	if mgr.move.onTile {
		// Update tile currently in motion.
		t := mgr.move.tile
		t.Loc = Sub(l, t.MoveOffset)
	}
}

func onMouseUp(event js.Value) {
	if !mgr.move.active {
		fmt.Println("saw a mouseUp with no active move recorded!")
		return
	}
	l := clickOffset(event)
	ci := clickInfo(l)

	// Drop tile here or send to tray if not possible.
	if mgr.move.onTile {
		t := mgr.move.tile
		switch ci.zone {
		case ZoneBoard:
			// Release tile onto board.
			t.addToBoard(ci.coords)
			if !mgr.highlight.active {
				highlightCoords(ci.coords)
			}
		case ZoneTray:
			// Release tile onto tray.
			t.addToTray(ci.coords)
		default:
			t.sendToTray()
		}

		// Mark all tiles as valid when a new tile is placed on or from the board.
		if mgr.move.startZone == ZoneBoard || t.Zone == ZoneBoard {
			markAllTilesValid()
		}
	}

	// Update highlight, if needed.
	switch ci.zone {
	case ZoneBoard:
		if mgr.move.startClick == l && !mgr.move.hasMoved {
			if mgr.highlight.active && mgr.highlight.idx == ci.coords {
				// click was on highlight, so toggle instead.
				toggleWordDir()
			} else {
				highlightCoords(ci.coords)
			}
		}
	default:
		unhighlight()
	}

	// Reset stored move info and listeners.
	mgr.move = &Move{}
	mgr.listens.EndMove()
}

func onKeyDown(event js.Value) {
	if !mgr.highlight.active {
		return
	}
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
}
