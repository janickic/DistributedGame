package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"encoding/gob"
)

// ClientManager holds available clients
type ClientManager struct {
	clients          []net.Conn
	lock             sync.RWMutex
	receive          chan Message
	disconnectClient chan net.Conn
	gameStarted      bool
}

func startServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", ":12345")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients:          make([]net.Conn, 0, 4),
		receive:          make(chan Message),
		disconnectClient: make(chan net.Conn),
		gameStarted:      false,
	}
	
	var players [4]Player

	game := Game{
		N: 4, //TODO: Make customizable
		MinFill: 0.6, //TODO: Make customizable
		Players: players,
		Active: false,
	}

	// start channels
	go manager.startChannels()

	for {
		if manager.gameStarted == false {
			connection, err := listener.Accept()
			fmt.Println("Client connected, ", len(manager.clients)+1)
			if err != nil {
				fmt.Println(err)
			}
			manager.lock.Lock()
			manager.clients = append(manager.clients, connection)
			gob.Register(Player{})
			player := Player{
				Id: int64(len(manager.clients)-1),
				Ip: connection.RemoteAddr().(*net.TCPAddr).IP,
				Colour: 5, //TODO: Make an actual colour
				Score: 0,
			}
			game.Players[len(manager.clients)-1] = player
			gob.Register(Player{})
			message := Message{
				MsgType: dataPlayer,
				Body: player,
			}
			gobEncoder := gob.NewEncoder(connection)
			err = gobEncoder.Encode(message)
			if err != nil {
				fmt.Println("encoding error: ", err)
			}

			if len(manager.clients) == 3 {
				manager.gameStarted = true
				game.Active = true
			}
			manager.lock.Unlock()

			// Start goroutine for listening on this client
			go manager.receiveMessages(connection)

			if len(manager.clients) == 3 {
				manager.startGame(game)
			}
		}

	}

}

func (manager *ClientManager) startGame(game Game) {
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Start Game")
	game.Active = true
	gob.Register(Game{})
	message := Message{
		MsgType: dataGame,
		Body: game,
	}
	// Send message to clients to start game
	for _, client := range manager.clients {
		gobEncoder := gob.NewEncoder(client)
		err := gobEncoder.Encode(message)
		if err != nil {
			fmt.Println("encoding error: ", err)
		}
	}
}

/*
	Start channels
*/
func (manager *ClientManager) startChannels() {
	for {
		select {
		// Receive message from channel
		case message := <-manager.receive:
			// Send message to all other clients
			for _, client := range manager.clients {
				gobEncoder := gob.NewEncoder(client)
				err := gobEncoder.Encode(message)
				if err != nil {
					fmt.Printf("Couldn't send message to client %+v\n", client)
				}
			}
			
		case connection := <-manager.disconnectClient:
			for index, client := range manager.clients {
				if client == connection {
					fmt.Println("Terminate this connection, ", index+1)
					connection.Close()
					manager.clients = append(manager.clients[:index], manager.clients[index+1:]...)
				}
			}

		}

	}
}

func ServerHandleMove(move Move, curCell *Cell) bool{
	curCell.Lock()
	defer curCell.Unlock()
	switch move.Action {
	case lock:
		if !curCell.Locked{
			curCell.Owner = move.Player
			curCell.Locked = true
			return true
		}
	case unlock:
		if curCell.Owner.Id == move.Player.Id{
			curCell.Owner = Player{}
			curCell.Locked = false
			return true
		}

	case fill:
		if !curCell.Locked || curCell.Owner.Id == move.Player.Id{
			curCell.Owner = move.Player
			curCell.Locked = true
			curCell.Filled = true
			return true
		}
	}
	fmt.Println("Move failed")
	return false
}

/*
	RECEIVE MESSAGES FROM CLIENTS
*/
func (manager *ClientManager) receiveMessages(client net.Conn) {
	for {
		message := &Message{}
		gobDecoder := gob.NewDecoder(client)
		err := gobDecoder.Decode(message)
		if err != nil {
			manager.disconnectClient <- client
			fmt.Println("Error in socket connection,", err)
			client.Close()
			break
		}
		switch message.MsgType{
		case dataGame:
			fmt.Println("Received Game")
			fmt.Println("should update player on game state")
		case dataPlayer:
			fmt.Println("for some reason server received a player")
		case dataMove:
			nextMove := message.Body.(Move)
			fmt.Println("received move")
			curCell := &curGame.Board[nextMove.CellX][nextMove.CellY]
			success := ServerHandleMove(nextMove, curCell)
			if success {
				gob.Register(Move{})
				nextMove.Timestamp = time.Now()
				acceptedMove := Message{
					MsgType: dataMove,
					Body: nextMove,
				}

				//gobEncoder := gob.NewEncoder(manager.receive)
				//err = gobEncoder.Encode(acceptedMove)
				manager.receive <- acceptedMove
			}
		}
		
	}
}
