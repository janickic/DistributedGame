package main

func createBlockArray(
	screenDim int,
	totalPixels int,
	blockDim int,
	percentColor float32) []block {

	numberOfBlocks := totalPixels / (blockDim * blockDim)
	blockArray := make([]block, numberOfBlocks)

	xOffset := 0
	yOffset := 0

	for i := 0; i < numberOfBlocks; i++ {
		blockArray[i] = initBlock(i, xOffset, yOffset, screenDim, blockDim, percentColor)

		xOffset = xOffset + blockDim

		if xOffset >= screenDim {
			xOffset = 0
			yOffset = yOffset + blockDim
		}
	}

	return blockArray
}

func choosePlayerColor(id int64) (color rgb_color) {
	switch id {
	case 0:
		return newColor(0, 0, 255)
	case 1:
		return newColor(0, 255, 0)
	case 2:
		return newColor(255, 0, 255)
	default:
		return newColor(0, 255, 255)
	}
}
