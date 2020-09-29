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
