package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

// ClientManager holds available clients
type ClientManager struct {
	clients          []net.Conn
	lock             sync.RWMutex
	receive          chan Message
	disconnectClient chan net.Conn
	gameStarted      bool
}

var serverGame = Game{}

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

	//will make nxn board
	n := 4
	var board [][]Cell
	for i := 0; i < n; i++ {
		board = append(board, make([]Cell, n))
	}

	var players [4]Player

	serverGame = Game{
		Board:   board,
		N:       n,
		MinFill: 0.6,
		Players: players,
		Active:  false,
		id:      0,
	}

	// start channels
	go manager.startChannels()

	for {
		if manager.gameStarted == false {
			connection, err := listener.Accept()
			fmt.Println("Client connected, ", len(manager.clients))
			if err != nil {
				fmt.Println(err)
			}
			manager.lock.Lock()
			manager.clients = append(manager.clients, connection)
			gob.Register(Player{})
			player := Player{
				Id:     int64(len(manager.clients) - 1),
				Ip:     connection.RemoteAddr().(*net.TCPAddr).IP,
				Colour: 5,
				Score:  0,
			}
			serverGame.Players[len(manager.clients)-1] = player
			gob.Register(Player{})
			message := Message{
				MsgType: dataPlayer,
				Body:    player,
			}
			gobEncoder := gob.NewEncoder(connection)
			err = gobEncoder.Encode(message)
			if err != nil {
				fmt.Println("encoding error: ", err)
			}

			numOfPlayers := 4

			if len(manager.clients) == numOfPlayers {
				manager.gameStarted = true
				serverGame.Active = true
			}
			manager.lock.Unlock()

			// Start goroutine for listening on this client
			go manager.receiveMessages(connection)

			if len(manager.clients) == numOfPlayers {
				manager.startGame(serverGame)
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
		Body:    game,
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

func ServerHandleMove(move Move, curCell *Cell) bool {
	curCell.Lock()
	defer curCell.Unlock()
	switch move.Action {
	case lock:
		if !curCell.Locked {
			curCell.Owner = move.Player
			curCell.Locked = true
			return true
		}
	case unlock:
		if curCell.Owner.Id == move.Player.Id {
			curCell.Owner = Player{}
			curCell.Locked = false
			return true
		}

	case fill:
		if !curCell.Locked || curCell.Owner.Id == move.Player.Id {
			curCell.Owner = move.Player
			curCell.Locked = true
			curCell.Filled = true
			serverGame.Players[move.Player.Id].IncreaseScore()
			return true
		}
	}
	fmt.Printf("Player %+v move failed\n", move.Player.Id)
	return false
}

/*
	RECEIVE MESSAGES FROM CLIENTS
*/
func (manager *ClientManager) receiveMessages(client net.Conn) {
	gob.Register(Game{})
	gob.Register(Player{})
	gob.Register(Move{})
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
		switch message.MsgType {
		case dataGame:
			fmt.Println("Received Game")
			fmt.Println("should update player on game state")
		case dataPlayer:
			fmt.Println("for some reason server received a player")
		case dataMove:
			nextMove := message.Body.(Move)
			curCell := &serverGame.Board[nextMove.CellX][nextMove.CellY]

			success := ServerHandleMove(nextMove, curCell)
			if success {
				nextMove.Timestamp = time.Now()
				acceptedMove := Message{
					MsgType: dataMove,
					Body:    nextMove,
				}

				manager.receive <- acceptedMove
			}

		}

	}
}

//startNewServer is called when client needs to create new server
func startNewServer(game *Game) {
	fmt.Println("\n\n\nStarting new server...")

	manager := ClientManager{
		clients:          make([]net.Conn, 0, 4),
		receive:          make(chan Message),
		disconnectClient: make(chan net.Conn),
		gameStarted:      false,
	}
	fmt.Println("num of players left in game:", game.numOfPlayers)
	//because player automatically leaves because server is disconnected
	playerCounter := 0

	for i := 0; i < len(game.Players); i++ {
		if game.Players[i].Ip != nil {
			playerCounter++
		}
		if game.Players[i].Ip.String() == "127.0.0.1" {
			game.Players[i].Ip = nil
		}
	}

	//making new request to server
	for i := 0; i < playerCounter; i++ {

		if game.Players[i].Ip != nil {
			//this sends off messages to clients
			ip := game.Players[i].Ip.String()
			ip = fmt.Sprintf("%s:54321", ip)

			//end this after clients have accepted
			connection, err := net.Dial("tcp", ip)
			if err != nil {
				fmt.Println(err)
			}
			if connection == nil {
				fmt.Println("something went wrong")
			}

		}
	}

	fmt.Println("\nList of Players still in game:")
	for i := 0; i < playerCounter; i++ {
		if game.Players[i].Ip != nil {
			fmt.Println("Player: ", game.Players[i])
		}
	}
	fmt.Println("")

	playerCounter--
	serverGame = *game
	serverGame.numOfPlayers = playerCounter

	//reseting players
	var players [4]Player
	serverGame.Players = players

	serverGame.Active = false

	//may need to make this more specific but we'll see
	serverGame.id++

	// start channels
	go manager.startChannels()
	fmt.Println(
		"Old Game: \nn: ", serverGame.N,
		"\nMinFill: ", serverGame.MinFill,
		"\nActive: ", serverGame.Active,
		"\nid: ", serverGame.id,
		"\nnum of players: ", serverGame.numOfPlayers)

	fmt.Println("Clients connected to server")

}
