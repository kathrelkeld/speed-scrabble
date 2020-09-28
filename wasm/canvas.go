package main

import (
	"syscall/js"
)

type Canvas js.Value

func (c Canvas) Call(field string, args ...interface{}) {
	js.Value(c).Call(field, args...)
}

func (c Canvas) Set(field string, value interface{}) {
	js.Value(c).Set(field, value)
}

func (c Canvas) Get(field string) interface{} {
	return js.Value(c).Get(field)
}

func (c Canvas) Size() Vec {
	x := js.Value(c).Get("width").Int()
	y := js.Value(c).Get("height").Int()
	return Vec{x, y}
}

func (c Canvas) Start() Vec {
	x := js.Value(c).Get("offsetLeft").Int()
	y := js.Value(c).Get("offsetTop").Int()
	return Vec{x, y}
}

func (c Canvas) End() Vec {
	x := js.Value(c).Get("offsetRight").Int()
	y := js.Value(c).Get("offsetBottom").Int()
	return Vec{x, y}
}

func (c Canvas) IsNull() bool {
	return js.Value(c).IsNull()
}
