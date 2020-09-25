package main

import (
	"fmt"
	"syscall/js"
)

var ctx js.Value // set during page setUp

func drawTile(t *TileLoc) {
	size := manager.tileSize
	fmt.Println("drawing tile", t.value)
	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", t.canvasLoc.X, t.canvasLoc.Y, size.X, size.Y)
	ctx.Set("fillStyle", "red")
	ctx.Set("font", fmt.Sprintf("%v", size.X)+"px Arial")
	ctx.Call("fillText", t.value, t.canvasLoc.X, t.canvasLoc.Y+manager.tileSize.Y)
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
	drawBoard()
	drawTray()
	drawTiles()
}
