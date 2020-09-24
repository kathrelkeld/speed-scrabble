package main

import (
	"fmt"
	"syscall/js"

	"github.com/kathrelkeld/speed-scrabble/game"
)

var ctx js.Value // set during page setUp

func drawTile(t *game.Tile, v game.Vec) {
	fmt.Println("drawing tile", t.Value)
	ctx.Call("fillRect", 10, 10, 100, 100)
}

func drawLineBetween(a, b game.Vec) {
	ctx.Call("beginPath")
	ctx.Call("moveTo", a.X, a.Y)
	ctx.Call("lineTo", b.X, b.Y)
	ctx.Call("stroke")
}

func drawGrid(start, gridSize, tileSize game.Vec) {
	end := game.Vec{
		gridSize.X*tileSize.X + start.X,
		gridSize.Y*tileSize.Y + start.Y,
	}
	for i := start.X; i <= end.X; i += tileSize.X {
		drawLineBetween(game.Vec{i, start.Y}, game.Vec{i, end.Y})
	}
	for j := start.Y; j <= end.Y; j += tileSize.Y {
		drawLineBetween(game.Vec{start.X, j}, game.Vec{end.X, j})
	}
}

func drawTray() {
	drawGrid(game.Vec{10, 450}, manager.traySize, manager.tileSize)
}

func drawBoard() {
	drawGrid(game.Vec{10, 10}, manager.boardSize, manager.tileSize)
}

func draw() {
	drawBoard()
	drawTray()
}
