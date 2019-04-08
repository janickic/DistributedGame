package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type block struct {
	pixels        []pixel
	isFilled      bool
	busy          bool
	owner         int
	coloredPixels int

	offsetX      int
	offsetY      int
	dimension    int
	blockID      int
	percentColor float32
}

//make block 50x50 -> 2500 pixels
//window Dimension
func initBlock(blockID int, xIndex int, yIndex int, windowD int, blockDim int, percentColor float32) (b block) {

	//Configure fields for Block
	b.dimension = blockDim
	b.percentColor = percentColor
	b.isFilled = false
	b.owner = -1
	b.coloredPixels = 0
	b.blockID = blockID
	b.offsetX = xIndex
	b.offsetY = yIndex

	b.pixels = createPixelArray(b.offsetX, b.offsetY, b.dimension)

	return b
}

// func createPixelArray(offsetX int, offsetY int, dimension int) []sdl.Rect {
func createPixelArray(offsetX int, offsetY int, dimension int) []pixel {
	numberOfBlockPixels := dimension * dimension
	bottomBorder := numberOfBlockPixels - (2 * dimension)
	pixelArray := make([]pixel, numberOfBlockPixels)
	xCoor := 0
	yCoor := 0

	for i := 0; i < numberOfBlockPixels; i++ {
		xWithOffset := int32(xCoor + offsetX)
		yWithOffset := int32(yCoor + offsetY)
		pixelNumber := xCoor + yCoor*dimension

		// if the pixel is regular pixel or border pixel
		// arbitrarily chose if border pixels are 2 pixels from edge
		canChange := true
		if pixelNumber < 2*dimension ||
			(pixelNumber+1)%dimension == 0 ||
			(pixelNumber+2)%dimension == 0 ||
			(pixelNumber)%dimension == 0 ||
			(pixelNumber-1)%dimension == 0 ||
			pixelNumber > bottomBorder {

			canChange = false
		}

		pixelArray[i] = newPixel(canChange, sdl.Rect{xWithOffset, yWithOffset, 1, 1})
		xCoor = xCoor + 1

		if xCoor >= dimension {
			xCoor = 0
			yCoor = yCoor + 1
		}
	}

	return pixelArray
}

//renderBlock draws boxes on screen, they are either white and in the middle, or black and a border
func (b *block) renderBlock(renderer *sdl.Renderer) {
	for i := 0; i < len(b.pixels); i++ {
		if b.pixels[i].canChange {
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.FillRect(&b.pixels[i].val)

		} else {
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.FillRect(&b.pixels[i].val)
		}
	}
}

//drawOnBlock determines if a user can draw on a block
func (b *block) drawOnBlock(
	renderer *sdl.Renderer,
	mouseX int,
	mouseY int,
	blockDim int,
	p *player,
	prevX int,
	prevY int) {

	blockIndex := (mouseX - b.offsetX) + (mouseY-b.offsetY)*blockDim

	if b.pixels[blockIndex].canChange {
		rgb := p.color

		renderer.SetDrawColor(rgb.r, rgb.g, rgb.b, 255)
		renderer.DrawRect(&b.pixels[blockIndex].val)
		renderer.FillRect(&b.pixels[blockIndex].val)

		if prevX != mouseX && prevY != mouseY {
			b.coloredPixels++
		}
		// b.coloredPixels++
	}
}

// isAllowed checks to see if block can be coloured - need to setup w/ network though
func (b *block) isAllowed(p *player) bool {
	if !b.isFilled && p.currentBlock == int32(b.blockID) {
		return true
	}

	return false
}

// blockFilled checks if minimum number of blocks are filled
func (b *block) blockFilled() bool {
	filled := float32(b.coloredPixels) / float32(b.dimension*b.dimension)

	if (filled * 100) >= b.percentColor {
		return true
	}

	return false
}

func (b *block) completeBlock(p *player, renderer *sdl.Renderer) {
	rgb := p.color

	b.fillBlock(rgb.r, rgb.g, rgb.b, renderer)
	b.coloredPixels = b.dimension * b.dimension
	b.isFilled = true
	b.owner = int(p.id)
}

func (b *block) resetBlock(renderer *sdl.Renderer) {
	b.fillBlock(255, 255, 255, renderer)
	b.coloredPixels = 0
	b.busy = false
	b.isFilled = false
	b.owner = -1
}

// fillBlock will either fill the block or undo the changes made by the player
func (b *block) fillBlock(red, green, blue uint8, renderer *sdl.Renderer) {
	for i := 0; i < len(b.pixels); i++ {
		if b.pixels[i].canChange {
			renderer.SetDrawColor(red, green, blue, 255)
			renderer.FillRect(&b.pixels[i].val)
		}
	}

}
