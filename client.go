package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// Client provides socket info and data for single client
type Client struct {
	socket net.Conn
	data   chan []byte
}

const (
	screenDim     = 600
	blockDim      = 100
	totalScreen   = screenDim * screenDim
	blocksPerPage = screenDim / blockDim
	percentColor  = 0.6
)

var curGame = Game{}
var myPlayer = Player{}

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

	// /*
	//  Other stuff
	// */
	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	message, _ := reader.ReadString('\n')
	// 	// Send data to server
	// 	connection.Write([]byte(strings.TrimRight(message, "\n")))
	// }

	////////// Begin of Mackenzie Frontend //////////////

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

	// Create New Player
	rgb := newColor(255, 0, 0)
	p := newPlayer(1, rgb)
	reloadScreen := 1

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

			//if the user has not touched a block yet
			if p.currentBlock == -1 {
				p.currentBlock = boxIndex
			}

			//if block is currently unfinished or not owned by anyone and if the user is currently writing on it
			if blockArray[boxIndex].isAllowed(&p) {
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

	////////// End of Mackenzie Frontend ///////////
}

/*
	RECEIVE MESSAGES From SERVER
*/
func (client *Client) socketReceive() {
	gob.Register(Game{})
	gob.Register(Player{})

	for {
		message := &Message{}
		gobDecoder := gob.NewDecoder(client.socket)
		err := gobDecoder.Decode(message)
		if err != nil {
			fmt.Println("decoding error: ", err)
		}
		switch message.MsgType {
		case dataGame:
			curGame = message.Body.(Game)
			fmt.Println("Received Game")
			//Test move
			//client.OnMouseDown(0, 0)
		case dataPlayer:
			myPlayer = message.Body.(Player)
			fmt.Println("I am player", myPlayer.Id)
		case dataMove:
			nextMove := message.Body.(Move)
			fmt.Println("received move")
			curCell := &curGame.Board[nextMove.CellX][nextMove.CellY]
			ClientHandleMove(nextMove, curCell, myPlayer.Id == nextMove.Player.Id)
		}

	}
}

func ClientHandleMove(move Move, curCell *Cell, isMe bool) {
	curCell.Lock()
	defer curCell.Unlock()
	switch move.Action {
	case lock:
		curCell.Owner = move.Player
		curCell.Locked = true
		if isMe {
			fmt.Println("Start drawing line")
		}
	case unlock:
		curCell.Owner = Player{}
		curCell.Locked = false
		if isMe {
			fmt.Println("Erase line")
		}
	case fill:
		curCell.Owner = move.Player
		curCell.Locked = true
		curCell.Filled = true
		curGame.Players[move.Player.Id].IncreaseScore()
		fmt.Println("this should update gui board")
	}
}

func (client *Client) OnMouseDown(cellX, cellY int) {
	gob.Register(Move{})
	curCell := &curGame.Board[cellX][cellY]
	if !curCell.Locked {
		move := Move{
			CellX:     cellX,
			CellY:     cellY,
			Action:    lock,
			Player:    myPlayer,
			Timestamp: time.Now(),
		}

		message := Message{
			MsgType: dataMove,
			Body:    move,
		}

		gobEncoder := gob.NewEncoder(client.socket)
		err := gobEncoder.Encode(message)
		if err != nil {
			fmt.Println("encoding error: ", err)
		}
	}
}

func (client *Client) OnMouseUp(cellX, cellY int, success bool) {
	gob.Register(Move{})
	move := Move{
		CellX:     cellX,
		CellY:     cellY,
		Player:    myPlayer,
		Timestamp: time.Now(),
	}

	if success {
		move.Action = fill
	} else {
		move.Action = unlock
	}

	message := Message{
		MsgType: dataMove,
		Body:    move,
	}

	gobEncoder := gob.NewEncoder(client.socket)
	err := gobEncoder.Encode(message)
	if err != nil {
		fmt.Println("encoding error: ", err)
	}
}

func (client *Client) chanReceive() {
	for {
		data := <-client.data
		fmt.Println("RECEIVED: " + string(data))
	}
}
