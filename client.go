package main

import (
	"fmt"
	"net"

	"github.com/veandco/go-sdl2/sdl"
)

// Client provides socket info and data for single client
type Client struct {
	socket net.Conn
	data   chan []byte
}

const (
	screenDim     = 600
	blockDim      = 150
	totalScreen   = screenDim * screenDim
	blocksPerPage = screenDim / blockDim
	percentColor  = 0.6
)

var curGame = Game{}
var myPlayer = Player{}

// Create New Player for Client
var rgb = newColor(255, 0, 0)
var p = newPlayer(myPlayer.Id, rgb)

func startClientMode(ip string) {
	fmt.Println("Starting client...")
	fmt.Println("Please wait for Start Game message from server")
	ip = fmt.Sprintf("%s:12345", ip)
	connection, error := net.Dial("tcp", ip)
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{
		socket: connection,
		data:   make(chan []byte),
	}

	go client.socketReceive()
	go client.chanReceive()

	//////--------- Begin of Mackenzie Frontend ----------//////

	for !curGame.Active {
	}
	fmt.Println("3 Clients connected!")

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Println("initializing SDL:", err)
		return
	}

	window, err := sdl.CreateWindow(
		"Game Start",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		screenDim,
		screenDim,
		sdl.WINDOW_OPENGL)

	if err != nil {
		fmt.Println("initializing SDL:", err)
		return
	}

	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		fmt.Println("initializing SDL:", err)
		return
	}

	defer renderer.Destroy()
	defer window.Destroy()

	blockArray := createBlockArray(
		screenDim,
		totalScreen,
		blockDim,
		percentColor)

	reloadScreen := 1
	mouseToServer := false

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch val := event.(type) {

			case *sdl.QuitEvent:
				return

			case *sdl.KeyboardEvent:
				if val.Keysym.Sym == sdl.K_SPACE {
					fmt.Println("Board Created")

					//creates board - just press spacebar when loaded up
					//pressing this twice works for some reason
					for i := 0; i < len(blockArray); i++ {
						blockArray[i].renderBlock(renderer)
					}

					renderer.Present()
				}

				if val.Keysym.Sym == sdl.K_RETURN {
					fmt.Println("Score is: ", p.score)
				}
			}
		}

		mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
		if mouseButtonState == 1 {
			boxIndex := (mouseX / blockDim) + (mouseY/blockDim)*blocksPerPage
			serverX := int(boxIndex % blocksPerPage)
			serverY := int((boxIndex / blocksPerPage) % blocksPerPage)

			//if the user has not touched a block yet
			if p.currentBlock == -1 {
				p.currentBlock = boxIndex
			}

			if !mouseToServer {
				client.OnMouseDown(serverX, serverY)
				mouseToServer = true
			}

			//if block is currently unfinished or not owned by anyone and if the user is currently writing on it
			// if blockArray[boxIndex].isAllowed(&p) {
			if p.canWrite && blockArray[boxIndex].isAllowed(&p) {
				p.active = true
				blockArray[boxIndex].drawOnBlock(renderer, int(mouseX), int(mouseY), blockDim, &p)

				//this thing is the issue, to many re-renders
				if reloadScreen%100 == 0 {

					renderer.Present()
					// fmt.Println("reload screen", reloadScreen)
					reloadScreen = 1

				} else {
					reloadScreen++
				}

			} else if p.currentBlock != boxIndex {
				//do nothing, they've gone out of bounds

			} else {
				p.currentBlock = -1
			}

			//when player unclicks
		} else {
			if p.active {
				mouseToServer = false
				p.disableWrite()

				if blockArray[p.currentBlock].blockFilled() {
					blockArray[p.currentBlock].completeBlock(&p, renderer)
					fmt.Println("You coloured all of it!")

				} else {
					blockArray[p.currentBlock].resetBlock(renderer)
					fmt.Println("You didn't colour all of it :(")
				}

				p.currentBlock = -1
				p.active = false
				renderer.Present()
			}

		}

		// renderer.Present()
	}

}
