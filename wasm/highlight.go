package main

type Highlight struct {
	active bool
	idx    Vec // Coordinates on the board where the highlight starts.
	dir    Vec // Direction to advance when typing.
}

// moveHighlight moves the highlight in the given direction.
func (mgr *GameManager) moveHighlight(d Vec) {
	if !mgr.highlight.active {
		return
	}
	newSpace := Add(mgr.highlight.idx, d)
	if mgr.board.InCoords(newSpace) {
		mgr.highlight.idx = newSpace
	}
}

// findForHighlight finds a tile of value v in the tray and moves it to the highlight.
// There is definitely a highlight before this gets called.
// Does nothing if there is already a matching tile present.
func (mgr *GameManager) findForHighlight(v string) {
	if prev := mgr.board.Get(mgr.highlight.idx); prev != nil && prev.Value == v {
		mgr.moveHighlight(mgr.highlight.dir)
		return
	}
	for _, t := range mgr.tiles {
		if t.Zone == ZoneTray && t.Value == v {
			t.addToBoard(mgr.highlight.idx)
			mgr.moveHighlight(mgr.highlight.dir)
			return
		}
	}
}

func (mgr *GameManager) toggleWordDir() {
	mgr.highlight.dir = Vec{mgr.highlight.dir.Y, mgr.highlight.dir.X}
}

func (mgr *GameManager) backspaceHighlight() {
	if t := mgr.board.Get(mgr.highlight.idx); t != nil {
		t.sendToTray()
	}
	mgr.moveHighlight(ScaleUp(mgr.highlight.dir, -1))
}

// highlightCoords highlights the square at location l, which is definitely on the board.
func (mgr *GameManager) highlightCoords(l Vec) {
	mgr.highlight.active = true
	mgr.highlight.idx = l
}

// highlightCanvas highlights the square at location l, which is definitely on the board.
func (mgr *GameManager) highlightCanvas(l Vec) {
	newH := mgr.board.coords(l)
	mgr.highlightCoords(newH)
}

// unhighlight removes all active highlights.
func (mgr *GameManager) unhighlight() {
	mgr.highlight.active = false
	mgr.highlight.idx = Vec{-1, -1}
	mgr.highlight.dir = Vec{1, 0}
}
