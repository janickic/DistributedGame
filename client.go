package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

// Client provides socket info and data for single client
type Client struct {
	socket net.Conn
	data   chan []byte
}

const (
	screenDim       = 600
	numberOfSquares = 4
	blockDim        = 600 / numberOfSquares
	totalScreen     = screenDim * screenDim
	blocksPerPage   = screenDim / blockDim
	percentColor    = 0.2
)

var mutex sync.RWMutex

var curGame = Game{}
var myPlayer = Player{}

// Create New Player for Client
var rgb = newColor(255, 0, 0)
var p = newPlayer(myPlayer.Id, rgb)
var renderer *sdl.Renderer
var blockMutex sync.Mutex
var blockArray = createBlockArray(
	screenDim,
	totalScreen,
	blockDim,
	percentColor)

var gameState = State{}

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
	go client.listenForServer()

	//////--------- Begin of Mackenzie Frontend ----------//////

	for !curGame.Active {
	}
	fmt.Println("NUM OF PLAYERS RIGHT NOW:", curGame.numOfPlayers)
	fmt.Println("Clients connected!")
	p.id = myPlayer.Id
	p.color = choosePlayerColor(p.id)

	//initializing game state
	gameState.blockArray = createBlockArray(
		screenDim,
		totalScreen,
		blockDim,
		percentColor)
	gameState.clientPlayer = p
	gameState.serverPlayer = myPlayer
	gameState.game = curGame
	curGame.restart = false

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

	renderer, err = sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		fmt.Println("initializing SDL:", err)
		return
	}

	gameState.renderer = renderer

	defer renderer.Destroy()
	defer window.Destroy()

	reloadScreen := 1
	mouseToServer := false

	prevX := int32(0)
	prevY := int32(0)

	for !curGame.restart {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch val := event.(type) {

			case *sdl.QuitEvent:
				return

			case *sdl.KeyboardEvent:
				if val.Keysym.Sym == sdl.K_SPACE {
					fmt.Println("Board Created")
					blockMutex.Lock()
					for i := 0; i < len(blockArray); i++ {
						blockArray[i].renderBlock(renderer)
					}
					blockMutex.Unlock()

					gameState.renderer.Present()
				}

				if val.Keysym.Sym == sdl.K_RETURN {
					fmt.Println("Score is: ", p.score)
				}
			}
		}

		mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
		boxIndex := (mouseX / blockDim) + (mouseY/blockDim)*blocksPerPage
		serverX := int(boxIndex % blocksPerPage)
		serverY := int((boxIndex / blocksPerPage) % blocksPerPage)

		if mouseButtonState == 1 {
			//if the user has not touched a block yet
			if p.currentBlock == -1 {
				p.currentBlock = boxIndex
			}

			if !mouseToServer {
				client.OnMouseDown(serverX, serverY)
				mouseToServer = true
			}

			//if block is currently unfinished or not owned by anyone and if the user is currently writing on it
			blockMutex.Lock()
			if p.canWrite && blockArray[boxIndex].isAllowed(&p) {
				p.active = true
				blockArray[boxIndex].drawOnBlock(
					renderer,
					int(mouseX),
					int(mouseY),
					blockDim,
					&gameState.clientPlayer,
					int(prevX),
					int(prevY))

				prevX = mouseX
				prevY = mouseY

				//this thing is the issue, to many re-renders
				if reloadScreen%100 == 0 {

					// renderer.Present()
					gameState.renderer.Present()
					reloadScreen = 1

				} else {
					reloadScreen++
				}

			} else if p.currentBlock != boxIndex {
				//do nothing, they've gone out of bounds

			} else {
				p.currentBlock = -1
			}
			blockMutex.Unlock()

		} else {

			//when player unclicks
			if p.active {
				mouseToServer = false
				p.disableWrite()

				blockMutex.Lock()
				blockWasFilled := blockArray[p.currentBlock].blockFilled()
				if blockWasFilled {
					blockArray[p.currentBlock].completeBlock(&p, renderer)
					p.score++
					fmt.Println("You coloured all of it!")

				} else {
					blockArray[p.currentBlock].resetBlock(renderer)
					fmt.Println("You didn't colour all of it :(")
				}
				blockMutex.Unlock()

				client.OnMouseUp(serverX, serverY, blockWasFilled)

				p.currentBlock = -1
				p.active = false
				renderer.Present()
			}
			mouseToServer = false
		}
	}

	//something

}
