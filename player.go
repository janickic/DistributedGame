package main

type player struct {
	id           int64
	currentBlock int32
	score        int
	active       bool
	color        rgb_color
	canWrite     bool
}

type rgb_color struct {
	r uint8
	g uint8
	b uint8
}

func newColor(r, g, b uint8) (rgb rgb_color) {
	rgb.r = r
	rgb.g = g
	rgb.b = b

	return rgb
}

func newPlayer(id int64, rgb rgb_color) (p player) {

	p.id = id
	p.currentBlock = -1
	p.score = 0
	p.active = false
	p.color = rgb
	p.canWrite = false

	return p
}

func (p *player) requestAccess() bool {
	return true
}

func (p *player) disableWrite() {
	p.canWrite = false
}

func (p *player) allowWrite() {
	p.canWrite = true
}
