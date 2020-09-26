package main

import ()

type Vec struct {
	X int
	Y int
}

func (a Vec) Add(b Vec) Vec {
	return Vec{a.X + b.X, a.Y + b.Y}
}

func (a Vec) Sub(b Vec) Vec {
	return Vec{a.X - b.X, a.Y - b.Y}
}

func (v Vec) ScaleUp(a int) Vec {
	return Vec{v.X * a, v.Y * a}
}

func (v Vec) ScaleDown(a int) Vec {
	return Vec{v.X / a, v.Y / a}
}

func cAdd(a, b canvasLoc) canvasLoc {
	return canvasLoc{a.X + b.X, a.Y + b.Y}
}
