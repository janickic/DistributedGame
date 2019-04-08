package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

// State is the Game state
type State struct {
	clientPlayer player
	serverPlayer Player
	renderer     *sdl.Renderer
	blockArray   []block
	game         Game
}
