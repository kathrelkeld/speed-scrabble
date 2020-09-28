package main

import (
	"fmt"
)

func (mgr *GameManager) drawTile(t *Tile) {
	// Draw box.
	switch t.State {
	case TileStateInvalid:
		mgr.ctx.Set("fillStyle", "red")
	default:
		mgr.ctx.Set("fillStyle", "black")
	}
	mgr.ctx.FillRect(t.Loc, mgr.tileSize)

	// Draw letter;
	switch t.State {
	case TileStateUnused:
		mgr.ctx.Set("fillStyle", "grey")
	default:
		mgr.ctx.Set("fillStyle", "white")
	}
	mgr.ctx.FillText(t.Value, Add(t.Loc, ScaleDown(mgr.tileSize, 2)))
}

func (mgr *GameManager) drawTiles() {
	//mgr.ctx.Set("globalAlpha", 1.0)
	mgr.ctx.Set("textAlign", "center")
	mgr.ctx.Set("textBaseline", "middle")
	mgr.ctx.Set("font", fmt.Sprintf("%v", mgr.tileSize.X)+"px Arial")
	for _, t := range mgr.tiles {
		mgr.drawTile(t)
	}
}

func (mgr *GameManager) drawBadWords() {
	if len(mgr.badWords) == 0 {
		return
	}
	mgr.ctx.BeginPath()
	mgr.ctx.Set("lineWidth", 2)
	mgr.ctx.Set("strokeStyle", "blue")
	for _, w := range mgr.badWords {
		tileCenter := ScaleDown(mgr.tileSize, 2)
		start := Add(mgr.board.canvasStart(w.Start), tileCenter)
		end := Add(mgr.board.canvasStart(w.End), tileCenter)
		mgr.drawLineBetween(start, end)
	}
	mgr.ctx.ClosePath()
	mgr.ctx.Stroke()
}

func (mgr *GameManager) drawWordDir() {
	if !mgr.highlight.active {
		return
	}
	start := mgr.board.canvasStart(Add(mgr.highlight.idx, mgr.highlight.dir))
	end := mgr.board.CanvasEnd()
	size := Vec{}
	if mgr.highlight.dir.X == 0 {
		size = Vec{mgr.tileSize.X, end.Y - start.Y}
	} else {
		size = Vec{end.X - start.X, mgr.tileSize.Y}
	}

	mgr.ctx.Set("globalAlpha", 0.4)
	mgr.ctx.Set("fillStyle", "yellow")
	mgr.ctx.FillRect(start, size)

	mgr.ctx.Set("globalAlpha", 1.0)
}

func (mgr *GameManager) drawHighlight() {
	if !mgr.highlight.active {
		return
	}
	l := mgr.board.canvasStart(mgr.highlight.idx)
	mgr.ctx.BeginPath()
	mgr.ctx.Set("strokeStyle", "yellow")
	mgr.ctx.Set("lineWidth", 8)
	mgr.ctx.Set("globalAlpha", 0.4)
	mgr.drawRectBetween(l, Add(l, mgr.tileSize), 4)
	mgr.ctx.ClosePath()
	mgr.ctx.Stroke()

	mgr.ctx.Set("globalAlpha", 1.0)
}

func (mgr *GameManager) drawRectBetween(a, b Vec, w int) {
	sides := [][]Vec{
		{{a.X - w, a.Y}, {b.X + w, a.Y}},
		{{b.X, a.Y - w}, {b.X, b.Y + w}},
		{{b.X + w, b.Y}, {a.X - w, b.Y}},
		{{a.X, b.Y + w}, {a.X, a.Y - w}},
	}
	for i := 0; i < len(sides); i++ {
		mgr.ctx.MoveTo(sides[i][0])
		mgr.ctx.LineTo(sides[i][1])
	}
}

func (mgr *GameManager) drawLineBetween(a, b Vec) {
	mgr.ctx.MoveTo(a)
	mgr.ctx.LineTo(b)
}

func (mgr *GameManager) drawGrid(g *Grid, tileSize Vec) {
	mgr.ctx.BeginPath()
	//mgr.ctx.Set("globalAlpha", 1.0)
	mgr.ctx.Set("lineWidth", 2)
	mgr.ctx.Set("strokeStyle", "black")
	end := g.CanvasEnd()
	for i := g.Loc.X; i <= end.X; i += tileSize.X {
		mgr.drawLineBetween(Vec{i, g.Loc.Y}, Vec{i, end.Y})
	}
	for j := g.Loc.Y; j <= end.Y; j += tileSize.Y {
		mgr.drawLineBetween(Vec{g.Loc.X, j}, Vec{end.X, j})
	}
	mgr.ctx.ClosePath()
	mgr.ctx.Stroke()
}

func (mgr *GameManager) draw() {
	if mgr.ctx.IsNull() || mgr.canvas.IsNull() {
		return
	}

	mgr.ctx.Clear(Vec{0, 0}, mgr.canvas.Size())
	if mgr.state != StatePlaying {
		return
	}

	mgr.drawWordDir()

	mgr.drawGrid(mgr.board, mgr.tileSize)
	mgr.drawGrid(mgr.tray, mgr.tileSize)
	mgr.drawTiles()

	mgr.drawBadWords()

	mgr.drawHighlight()
	if mgr.move.active && mgr.move.onTile {
		// Moving tiles should be on top.
		mgr.drawTile(mgr.move.tile)
	}
}
