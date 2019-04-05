package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type pixel struct {
	canChange bool
	val       sdl.Rect
}

func newPixel(canChange bool, value sdl.Rect) (p pixel) {
	p.canChange = canChange
	p.val = value

	return p
}
