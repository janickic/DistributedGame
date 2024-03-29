package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

func (client *Client) listenForServer() {
	fmt.Println("listen for server")
	listener, err := net.Listen("tcp", ":54321")
	if err != nil {
		fmt.Println(err)
	}

	//connection is waiting for this to work
	connection, err := listener.Accept()
	fmt.Println("\nServer restarting connected with me")
	if err != nil {
		fmt.Println(err)
	}
	if connection != nil {
		fmt.Println("here is IP of new server: ", connection.RemoteAddr().(*net.TCPAddr).IP)
	}

}

func (client *Client) socketReceive() {
	gob.Register(Game{})
	gob.Register(Player{})
	gob.Register(Move{})
	for {
		message := &Message{}
		gobDecoder := gob.NewDecoder(client.socket)
		err := gobDecoder.Decode(message)
		if err != nil {
			fmt.Println("decoding error: ", err)
		}

		switch message.MsgType {
		case dataGame:
			mutex.Lock()
			curGame = message.Body.(Game)
			mutex.Unlock()
			fmt.Println("Received Game")
		case dataPlayer:
			myPlayer = message.Body.(Player)
			fmt.Println("I am player", myPlayer.Id)
		case dataMove:
			nextMove := message.Body.(Move)
			mutex.Lock()
			curCell := &curGame.Board[nextMove.CellX][nextMove.CellY]
			mutex.Unlock()
			ClientHandleMove(nextMove, curCell, myPlayer.Id == nextMove.Player.Id)
		case dataError:
			if (myPlayer.Id - curGame.id) == 1 {
				fmt.Println("server has left the game, I am new host")
				curGame.id = myPlayer.Id
				go startNewServer(&curGame)
				for {

				}

			} else {
				fmt.Println("Just waiting for server, not new host")
				for {

				}
			}
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
			p.allowWrite()

		}
	case unlock:
		curCell.Owner = Player{}
		curCell.Locked = false
		if isMe {
			fmt.Println("Block was not written on")
		}
	case fill:
		curCell.Owner = move.Player
		curCell.Locked = true
		curCell.Filled = true

		mutex.Lock()
		curGame.Players[move.Player.Id].IncreaseScore()
		mutex.Unlock()
		index := move.CellX + (move.CellY * blocksPerPage)
		fillerPlayer := newPlayer(move.Player.Id, choosePlayerColor(move.Player.Id))

		blockMutex.Lock()
		blockArray[index].completeBlock(&fillerPlayer, gameState.renderer)
		blockMutex.Unlock()
		gameState.renderer.Present()
	}
}

// func (client *Client) OnMouseDown(cellX, cellY int, p *player) {
func (client *Client) OnMouseDown(cellX, cellY int) {
	gob.Register(Move{})
	mutex.Lock()
	curCell := &curGame.Board[cellX][cellY]
	mutex.Unlock()
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
			fmt.Println("encoding error for MouseDown: ", err)
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
