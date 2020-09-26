package main

import ()

type Vec struct {
	X int
	Y int
}

func Add(a, b Vec) Vec {
	return Vec{a.X + b.X, a.Y + b.Y}
}

func Sub(a, b Vec) Vec {
	return Vec{a.X - b.X, a.Y - b.Y}
}

func ScaleUp(v Vec, a int) Vec {
	return Vec{v.X * a, v.Y * a}
}

func ScaleDown(v Vec, a int) Vec {
	return Vec{v.X / a, v.Y / a}
}

func Mult(a, b Vec) Vec {
	return Vec{a.X * b.X, a.Y * b.Y}
}
