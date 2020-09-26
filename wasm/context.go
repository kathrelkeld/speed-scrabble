package main

import (
	"syscall/js"
)

type Context js.Value // 2D canvas context

func (c Context) Call(field string, args ...interface{}) {
	js.Value(c).Call(field, args...)
}

func (c Context) Set(field string, value interface{}) {
	js.Value(c).Set(field, value)
}

func (c Context) FillRect(l, size Vec) {
	js.Value(c).Call("fillRect", l.X, l.Y, size.X, size.Y)
}

func (c Context) FillText(v string, l Vec) {
	js.Value(c).Call("fillText", v, l.X, l.Y)
}

func (c Context) MoveTo(l Vec) {
	js.Value(c).Call("moveTo", l.X, l.Y)
}

func (c Context) LineTo(l Vec) {
	js.Value(c).Call("lineTo", l.X, l.Y)
}

func (c Context) Stroke() {
	js.Value(c).Call("stroke")
}

func (c Context) BeginPath() {
	js.Value(c).Call("beginPath")
}

func (c Context) ClosePath() {
	js.Value(c).Call("closePath")
}
