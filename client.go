package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"time"
	"strings"
)

// Client provides socket info and data for single client
type Client struct {
	socket net.Conn
	data   chan []byte
}

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

	/*
	 Other stuff
	*/
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		// Send data to server
		connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
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
		switch message.MsgType{
		case dataGame:
			curGame = message.Body.(Game)
			fmt.Println("Received Game")
		case dataPlayer:
			myPlayer = message.Body.(Player)
			fmt.Println("I am player", myPlayer.Id)
		case dataMove:
			nextMove := message.Body.(Move)
			fmt.Println("received moce")
			curCell := &curGame.Board[nextMove.CellX][nextMove.CellY]
			ClientHandleMove(nextMove, curCell, myPlayer.Id == nextMove.Player.Id)
		}
		
	}
}

func ClientHandleMove(move Move, curCell *Cell, isMe bool){
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
		//TODO: Update player scores
		fmt.Println("this should update gui board")
	}
}

func (client *Client) OnMouseDown(cellX, cellY int){
	gob.Register(Move{})
	curCell := &curGame.Board[cellX][cellY]
	if !curCell.Locked{
		move := Move{
			CellX: cellX,
			CellY: cellY,
			Action: lock,
			Player: myPlayer,
			Timestamp: time.Now(),
		}

		message := Message{
			MsgType: dataMove,
			Body: move,
		}

		gobEncoder := gob.NewEncoder(client.socket)
		err := gobEncoder.Encode(message)
		if err != nil {
			fmt.Println("encoding error: ", err)
		}
	}
}

func (client *Client) OnMouseUp(cellX, cellY int, success bool){
	gob.Register(Move{})
	move := Move{
		CellX: cellX,
		CellY: cellY,
		Player: myPlayer,
		Timestamp: time.Now(),
	}

	if success {
		move.Action = fill
	} else{
		move.Action = unlock
	}

	message := Message{
		MsgType: dataMove,
		Body: move,
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
