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
		x := event.Get("pageX").Int() - left
		y := event.Get("pageY").Int() - top
		// TODO do stuff here
		fmt.Println("x, y", x, y)
		return nil
	})
}

func inBounds(x, y, centerX, centerY, radius int) bool {
	return ((x < centerX+radius) && (x > centerX-radius) &&
		(y < centerY+radius) && (y > centerY-radius))
}

func setUpPage() {
	body := js.Global().Get("document").Get("body")

	canvas := js.Global().Get("document").Call("createElement", "canvas")
	// TODO tie canvas size to default game size
	canvas.Set("id", "canvas")
	canvas.Set("width", 1000)
	canvas.Set("height", 500)
	left := canvas.Get("offsetLeft").Int() + canvas.Get("clientLeft").Int()
	top := canvas.Get("offsetTop").Int() + canvas.Get("clientTop").Int()
	canvas.Call("addEventListener", "click", canvasOnClick(left, top))
	body.Call("appendChild", canvas)
}
