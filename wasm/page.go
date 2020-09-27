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
	jsFuncOf := func(f func()) js.Func {
		return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			f()
			return nil
		})
	}

	initializeListeners()
	body := js.Global().Get("document").Get("body")
	js.Global().Get("document").Call("addEventListener", "keydown", listenerKeyDown())

	// Add game buttons
	body.Call("appendChild", newButton("Reset Tiles", "reset", jsFuncOf(sendAllTilesToTray)))
	body.Call("appendChild", newButton("+1 Tile", "addTile", jsFuncOf(requestNewTile)))
	body.Call("appendChild", newButton("NewGame", "newGame", jsFuncOf(newGame)))
	body.Call("appendChild", newButton("Verify", "verify", jsFuncOf(verify)))
	body.Call("appendChild", newButton("Shuffle Tiles", "shuffle", jsFuncOf(shuffleTiles)))

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
