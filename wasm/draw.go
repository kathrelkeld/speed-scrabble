package main

import (
	"fmt"
	"syscall/js"
)

var canvas js.Value // set during page setUp
var ctx Context     // set during page setUp

func drawTile(t *TileLoc) {
	ctx.Set("fillStyle", "black")
	ctx.FillRect(t.canvasLoc, mgr.tileSize)
	if t.invalid {
		ctx.Set("fillStyle", "red")
	} else {
		ctx.Set("fillStyle", "white")
	}
	ctx.FillText(t.Value, canvasLoc(Vec(t.canvasLoc).Add(Vec(mgr.tileSize).ScaleDown(2))))
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
	if mgr.highlight == nil {
		return
	}
	start := mgr.board.canvasStart(gAdd(*mgr.highlight, mgr.wordDir))
	size := sizeV{}
	if mgr.wordDir.X == 0 {
		size = sizeV{mgr.tileSize.X, mgr.board.end.Y - start.Y}
	} else {
		size = sizeV{mgr.board.end.X - start.X, mgr.tileSize.Y}
	}

	ctx.Set("globalAlpha", 0.4)
	ctx.Set("fillStyle", "yellow")
	ctx.FillRect(start, size)

	ctx.Set("globalAlpha", 1.0)
}

func drawHighlight() {
	if mgr.highlight == nil {
		return
	}
	l := mgr.board.canvasStart(*mgr.highlight)
	ctx.BeginPath()
	ctx.Set("strokeStyle", "yellow")
	ctx.Set("lineWidth", 8)
	ctx.Set("globalAlpha", 0.4)
	drawRectBetween(l, cAdd(l, canvasLoc(mgr.tileSize)), 4)
	ctx.ClosePath()
	ctx.Stroke()

	ctx.Set("globalAlpha", 1.0)
}

func drawRectBetween(a, b canvasLoc, w int) {
	sides := [][]canvasLoc{
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

func drawLineBetween(a, b canvasLoc) {
	ctx.MoveTo(a)
	ctx.LineTo(b)
}

func drawGrid(g *Grid, tileSize sizeV) {
	ctx.BeginPath()
	//ctx.Set("globalAlpha", 1.0)
	ctx.Set("lineWidth", 2)
	ctx.Set("strokeStyle", "black")
	for i := g.loc.X; i <= g.end.X; i += tileSize.X {
		drawLineBetween(canvasLoc{i, g.loc.Y}, canvasLoc{i, g.end.Y})
	}
	for j := g.loc.Y; j <= g.end.Y; j += tileSize.Y {
		drawLineBetween(canvasLoc{g.loc.X, j}, canvasLoc{g.end.X, j})
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
	if mgr.movingTile != nil {
		// Moving tiles should be on top.
		drawTile(mgr.movingTile)
	}
}
