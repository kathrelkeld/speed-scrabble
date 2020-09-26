package main

import (
	"fmt"
	"syscall/js"
)

var canvas js.Value // set during page setUp
var ctx js.Value    // set during page setUp

func drawTile(t *TileLoc) {
	size := manager.tileSize
	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", t.canvasLoc.X, t.canvasLoc.Y, size.X, size.Y)
	if t.invalid {
		ctx.Set("fillStyle", "red")
	} else {
		ctx.Set("fillStyle", "white")
	}
	ctx.Set("textAlign", "center")
	ctx.Set("textBaseline", "middle")
	ctx.Set("font", fmt.Sprintf("%v", size.X)+"px Arial")
	ctx.Call("fillText", t.Value, t.canvasLoc.X+size.X/2,
		t.canvasLoc.Y+size.Y/2)
}

func drawTiles() {
	for _, t := range manager.tiles {
		drawTile(t)
	}
}

func drawLineBetween(a, b canvasLoc) {
	ctx.Call("beginPath")
	ctx.Call("moveTo", a.X, a.Y)
	ctx.Call("lineTo", b.X, b.Y)
	ctx.Call("stroke")
}

func drawGrid(start canvasLoc, gridSize, tileSize sizeV) {
	ctx.Set("fillStyle", "black")
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
