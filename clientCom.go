package main

import (
	"encoding/gob"
	"fmt"
	"time"
)

/*
	RECEIVE MESSAGES From SERVER
*/
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

		curGame.Players[move.Player.Id].IncreaseScore()
		index := move.CellX + (move.CellY * blocksPerPage)
		fillerPlayer := newPlayer(move.Player.Id, choosePlayerColor(move.Player.Id))

		blockArray[index].completeBlock(&fillerPlayer, gameState.renderer)

		gameState.renderer.Present()
	}
}

// func (client *Client) OnMouseDown(cellX, cellY int, p *player) {
func (client *Client) OnMouseDown(cellX, cellY int) {
	gob.Register(Move{})
	curCell := &curGame.Board[cellX][cellY]
	// if !curCell.Locked || p.id == curCell.Owner.Id {
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

func (client *Client) chanReceive() {
	for {
		data := <-client.data
		fmt.Println("RECEIVED: " + string(data))
	}
}
