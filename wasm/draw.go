package main

import (
	"fmt"
	"syscall/js"
)

var canvas js.Value // set during page setUp
var ctx Context     // set during page setUp

func drawTile(t *TileLoc) {
	ctx.Set("fillStyle", "black")
	ctx.FillRect(t.canvasLoc, manager.tileSize)
	if t.invalid {
		ctx.Set("fillStyle", "red")
	} else {
		ctx.Set("fillStyle", "white")
	}
	ctx.FillText(t.Value, canvasLoc(Vec(t.canvasLoc).Add(Vec(manager.tileSize).ScaleDown(2))))
	drawHighlight(t.canvasLoc)
}

func drawTiles() {
	//ctx.Set("globalAlpha", 1.0)
	ctx.Set("textAlign", "center")
	ctx.Set("textBaseline", "middle")
	ctx.Set("font", fmt.Sprintf("%v", manager.tileSize.X)+"px Arial")
	for _, t := range manager.tiles {
		drawTile(t)
	}
}

func drawHighlight(l canvasLoc) {
	ctx.BeginPath()
	ctx.Set("strokeStyle", "yellow")
	ctx.Set("lineWidth", 8)
	ctx.Set("globalAlpha", 0.4)
	drawRectBetween(l, cAdd(l, canvasLoc(manager.tileSize)), 4)
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

func drawGrid(start canvasLoc, gridSize, tileSize sizeV) {
	ctx.BeginPath()
	//ctx.Set("globalAlpha", 1.0)
	ctx.Set("lineWidth", 2)
	ctx.Set("strokeStyle", "black")
	end := canvasLoc{
		gridSize.X*tileSize.X + start.X,
		gridSize.Y*tileSize.Y + start.Y,
	}
	for i := start.X; i <= end.X; i += tileSize.X {
		drawLineBetween(canvasLoc{i, start.Y}, canvasLoc{i, end.Y})
	}
	for j := start.Y; j <= end.Y; j += tileSize.Y {
		drawLineBetween(canvasLoc{start.X, j}, canvasLoc{end.X, j})
	}
	ctx.ClosePath()
	ctx.Stroke()
}

func drawTray() {
	drawGrid(manager.trayLoc, manager.traySize, manager.tileSize)
}

func drawBoard() {
	drawGrid(manager.boardLoc, manager.boardSize, manager.tileSize)
}

func draw() {
	ctx.Call("clearRect", 0, 0, canvas.Get("width"), canvas.Get("height"))
	drawBoard()
	drawTray()
	drawTiles()
	if manager.movingTile != nil {
		// Moving tiles should be on top.
		drawTile(manager.movingTile)
	}
}
