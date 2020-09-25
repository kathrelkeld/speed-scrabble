package main

import (
	"fmt"
	"syscall/js"
)

func newButton(name, id string, onclick js.Func) js.Value {
	b := js.Global().Get("document").Call("createElement", "button")
	b.Set("innerHTML", name)
	b.Set("id", name)
	b.Set("onclick", onclick)
	return b
}

func canvasOnClick(left, top int) js.Func {
	// args = event
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		x := event.Get("offsetX").Int() - left
		y := event.Get("offsetY").Int() - top
		// TODO do stuff here
		fmt.Println("x, y", x, y)
		if t := onTile(canvasLoc{x, y}); t != nil {
			fmt.Println("on tile", t.value)
		}
		return nil
	})
}

func setUpPage() {
	body := js.Global().Get("document").Get("body")

	// Add game buttons
	body.Call("appendChild", newButton("Reset Tiles", "reset", sendAllTilesToTray()))
	body.Call("appendChild", newButton("+1 Tile", "addTile", requestNewTile()))
	body.Call("appendChild", newButton("NewGame", "reload", reload()))
	body.Call("appendChild", newButton("Verify", "verify", verify()))

	messages := js.Global().Get("document").Call("createElement", "textbox")
	messages.Set("id", "messages")
	body.Call("appendChild", messages)

	canvas := js.Global().Get("document").Call("createElement", "canvas")
	// TODO tie canvas size to default game size
	canvas.Set("id", "canvas")
	canvas.Set("width", 1000)
	canvas.Set("height", 500)
	left := canvas.Get("offsetLeft").Int() + canvas.Get("clientLeft").Int()
	top := canvas.Get("offsetTop").Int() + canvas.Get("clientTop").Int()
	canvas.Call("addEventListener", "click", canvasOnClick(left, top))
	body.Call("appendChild", canvas)
	ctx = canvas.Call("getContext", "2d")
}
