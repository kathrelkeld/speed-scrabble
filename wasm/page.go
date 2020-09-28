package main

import (
	"syscall/js"
)

type Listeners map[string]js.Func

func NewListeners(mgr *GameManager) Listeners {
	jsEventFuncOf := func(f func(js.Value)) js.Func {
		return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			f(args[0])
			args[0].Call("preventDefault")
			mgr.draw()
			return nil
		})
	}

	return Listeners{
		"mouseup":   jsEventFuncOf(mgr.onMouseUp),
		"mousemove": jsEventFuncOf(mgr.onMouseMove),
		"mousedown": jsEventFuncOf(mgr.onMouseDown),
		"keydown":   jsEventFuncOf(mgr.onKeyDown),
	}
}

func (l Listeners) addListener(where js.Value, which string) {
	where.Call("addEventListener", which, l[which])
}

func (l Listeners) removeListener(where js.Value, which string) {
	where.Call("removeEventListener", which, l[which])
}

func (l Listeners) NewGame() {
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	l.addListener(canvas, "mousedown")
	doc := js.Global().Get("document")
	l.addListener(doc, "keydown")
}

func (l Listeners) InMove() {
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	l.addListener(canvas, "mousemove")
	l.addListener(canvas, "mouseup")
}
func (l Listeners) EndMove() {
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	l.removeListener(canvas, "mousemove")
	l.removeListener(canvas, "mouseup")
}

func (l *Listeners) EndGame() {
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	l.removeListener(canvas, "mousedown")
	l.removeListener(canvas, "mousemove")
	l.removeListener(canvas, "mousemove")
	doc := js.Global().Get("document")
	l.removeListener(doc, "keydown")
}

func disableButton(id string) {
	js.Global().Get("document").Call("getElementById", id).Set("disabled", true)
}

func enableButton(id string) {
	js.Global().Get("document").Call("getElementById", id).Set("disabled", false)
}

func DisableGameButtons() {
	disableButton("resetTiles")
	disableButton("addTile")
	disableButton("verify")
	disableButton("shuffleTiles")
}

func EnableGameButtons() {
	enableButton("resetTiles")
	enableButton("addTile")
	enableButton("verify")
	enableButton("shuffleTiles")
}

func newButton(name, id string, onclick js.Func) js.Value {
	b := js.Global().Get("document").Call("createElement", "button")
	b.Set("innerHTML", name)
	b.Set("id", id)
	b.Set("onclick", onclick)
	return b
}

// jsFuncOf takes a function with no inputs and returns a js.Func that calls it.
func jsFuncOf(f func(), mgr *GameManager) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		f()
		mgr.draw()
		return nil
	})
}

func (mgr *GameManager) setUpPage() {
	body := js.Global().Get("document").Get("body")

	// Add game buttons
	body.Call("appendChild", newButton("Reset Tiles", "resetTiles", jsFuncOf(mgr.sendAllTilesToTray, mgr)))
	body.Call("appendChild", newButton("+1 Tile", "addTile", jsFuncOf(mgr.requestNewTile, mgr)))
	body.Call("appendChild", newButton("NewGame", "newGame", js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			mgr.newGame()
			return nil
		})))
	body.Call("appendChild", newButton("Verify", "verify", jsFuncOf(mgr.verify, mgr)))
	body.Call("appendChild", newButton("Shuffle Tiles", "shuffleTiles", jsFuncOf(mgr.shuffleTiles, mgr)))
	DisableGameButtons()

	messages := js.Global().Get("document").Call("createElement", "textbox")
	messages.Set("id", "messages")
	body.Call("appendChild", messages)

	canvas := js.Global().Get("document").Call("createElement", "canvas")
	// TODO tie canvas size to default game size
	canvas.Set("id", "canvas")
	canvas.Set("width", 1000)
	canvas.Set("height", 1000)
	body.Call("appendChild", canvas)

	mgr.ctx = Context(canvas.Call("getContext", "2d"))
	mgr.canvas = Canvas(canvas)
	mgr.listens = NewListeners(mgr)
}
