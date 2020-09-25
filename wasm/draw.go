package main

import (
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var ctx js.Value // set during page setUp

func drawTile(t *game.Tile, v canvasLoc) {
	fmt.Println("drawing tile", t.Value)
	ctx.Call("fillRect", 10, 10, 100, 100)
}

func drawLineBetween(a, b canvasLoc) {
	ctx.Call("beginPath")
	ctx.Call("moveTo", a.X, a.Y)
	ctx.Call("lineTo", b.X, b.Y)
	ctx.Call("stroke")
}

func drawGrid(start canvasLoc, gridSize, tileSize sizeV) {
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
	drawGrid(canvasLoc{10, 450}, manager.traySize, manager.tileSize)
}

func drawBoard() {
	drawGrid(canvasLoc{10, 10}, manager.boardSize, manager.tileSize)
}

func draw() {
	drawBoard()
	drawTray()
}
