package main

import (
	"syscall/js"
)

func newButton(name, id string, onclick js.Func) js.Value {
	b := js.Global().Get("document").Call("createElement", "button")
	b.Set("innerHTML", name)
	b.Set("id", name)
	b.Set("onclick", onclick)
	return b
}

func setUpPage() {
	initializeListeners()
	body := js.Global().Get("document").Get("body")
	js.Global().Get("document").Call("addEventListener", "keydown", listenerKeyDown())

	// Add game buttons
	body.Call("appendChild", newButton("Reset Tiles", "reset", sendAllTilesToTray()))
	body.Call("appendChild", newButton("+1 Tile", "addTile", requestNewTile()))
	body.Call("appendChild", newButton("NewGame", "newGame", newGame()))
	body.Call("appendChild", newButton("Verify", "verify", verify()))
	body.Call("appendChild", newButton("Shuffle Tiles", "shuffle", shuffleTiles()))

	messages := js.Global().Get("document").Call("createElement", "textbox")
	messages.Set("id", "messages")
	body.Call("appendChild", messages)

	canvas = js.Global().Get("document").Call("createElement", "canvas")
	// TODO tie canvas size to default game size
	canvas.Set("id", "canvas")
	canvas.Set("width", 1000)
	canvas.Set("height", 1000)
	canvas.Call("addEventListener", "mousedown", listenerMouseDown())
	body.Call("appendChild", canvas)
	ctx = Context(canvas.Call("getContext", "2d"))

}
