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

func draw() {

}
