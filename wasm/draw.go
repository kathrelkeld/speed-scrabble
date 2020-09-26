package main

import (
	"fmt"
	"syscall/js"
)

var canvas js.Value // set during page setUp
var ctx Context     // set during page setUp

func drawTile(t *Tile) {
	ctx.Set("fillStyle", "black")
	ctx.FillRect(t.Loc, mgr.tileSize)
	if t.Invalid {
		ctx.Set("fillStyle", "red")
	} else {
		ctx.Set("fillStyle", "white")
	}
	ctx.FillText(t.Value, Add(t.Loc, ScaleDown(mgr.tileSize, 2)))
}

func drawTiles() {
	//ctx.Set("globalAlpha", 1.0)
	ctx.Set("textAlign", "center")
	ctx.Set("textBaseline", "middle")
	ctx.Set("font", fmt.Sprintf("%v", mgr.tileSize.X)+"px Arial")
	for _, t := range mgr.tiles {
		drawTile(t)
	}
}

func drawWordDir() {
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

	ctx.Set("globalAlpha", 0.4)
	ctx.Set("fillStyle", "yellow")
	ctx.FillRect(start, size)

	ctx.Set("globalAlpha", 1.0)
}

func drawHighlight() {
	if !mgr.highlight.active {
		return
	}
	l := mgr.board.canvasStart(mgr.highlight.idx)
	ctx.BeginPath()
	ctx.Set("strokeStyle", "yellow")
	ctx.Set("lineWidth", 8)
	ctx.Set("globalAlpha", 0.4)
	drawRectBetween(l, Add(l, mgr.tileSize), 4)
	ctx.ClosePath()
	ctx.Stroke()

	ctx.Set("globalAlpha", 1.0)
}

func drawRectBetween(a, b Vec, w int) {
	sides := [][]Vec{
		{{a.X - w, a.Y}, {b.X + w, a.Y}},
		{{b.X, a.Y - w}, {b.X, b.Y + w}},
		{{b.X + w, b.Y}, {a.X - w, b.Y}},
		{{a.X, b.Y + w}, {a.X, a.Y - w}},
	}
	for i := 0; i < len(sides); i++ {
		ctx.MoveTo(sides[i][0])
		ctx.LineTo(sides[i][1])
	}
}

func drawLineBetween(a, b Vec) {
	ctx.MoveTo(a)
	ctx.LineTo(b)
}

func drawGrid(g *Grid, tileSize Vec) {
	ctx.BeginPath()
	//ctx.Set("globalAlpha", 1.0)
	ctx.Set("lineWidth", 2)
	ctx.Set("strokeStyle", "black")
	end := g.CanvasEnd()
	for i := g.Loc.X; i <= end.X; i += tileSize.X {
		drawLineBetween(Vec{i, g.Loc.Y}, Vec{i, end.Y})
	}
	for j := g.Loc.Y; j <= end.Y; j += tileSize.Y {
		drawLineBetween(Vec{g.Loc.X, j}, Vec{end.X, j})
	}
	ctx.ClosePath()
	ctx.Stroke()
}

func draw() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width"), canvas.Get("height"))

	drawWordDir()

	drawGrid(mgr.board, mgr.tileSize)
	drawGrid(mgr.tray, mgr.tileSize)
	drawTiles()

	drawHighlight()
	if mgr.move.active && mgr.move.onTile {
		// Moving tiles should be on top.
		drawTile(mgr.move.tile)
	}
}
