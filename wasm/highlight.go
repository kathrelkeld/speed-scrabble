package main

type Highlight struct {
	active bool
	idx    Vec // Coordinates on the board where the highlight starts.
	dir    Vec // Direction to advance when typing.
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

// highlightCoords highlights the square at location l, which is definitely on the board.
func highlightCoords(l Vec) {
	mgr.highlight.active = true
	mgr.highlight.idx = l
	draw()
}

// highlightCanvas highlights the square at location l, which is definitely on the board.
func highlightCanvas(l Vec) {
	newH := mgr.board.coords(l)
	highlightCoords(newH)
}

// unhighlight removes all active highlights.
func unhighlight() {
	mgr.highlight.active = false
	mgr.highlight.idx = Vec{-1, -1}
	mgr.highlight.dir = Vec{1, 0}
	draw()
}
