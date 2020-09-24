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

func drawGrid() {
	fmt.Println("drawing grid")
	gridStart := game.Vec{0, 0}
	gridEnd := game.Vec{
		manager.gridSize.X*manager.tileSize + gridStart.X,
		manager.gridSize.Y*manager.tileSize + gridStart.Y,
	}
	for i := gridStart.X; i <= gridEnd.X; i += manager.tileSize {
		drawLineBetween(game.Vec{i, gridStart.Y}, game.Vec{i, gridEnd.Y})
	}
	for j := gridStart.Y; j <= gridEnd.Y; j += manager.tileSize {
		drawLineBetween(game.Vec{gridStart.X, j}, game.Vec{gridEnd.X, j})
	}
}

func draw() {
	drawGrid()
}
