package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// ClientManager holds available clients
type ClientManager struct {
	clients          []net.Conn
	lock             sync.RWMutex
	receive          chan []byte
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
		receive:          make(chan []byte),
		disconnectClient: make(chan net.Conn),
		gameStarted:      false,
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
			if len(manager.clients) == 3 {
				manager.gameStarted = true
			}
			manager.lock.Unlock()

			// Start goroutine for listening on this client
			go manager.receiveMessages(connection)

			// send num clients connected
			for _, client := range manager.clients {
				numClients := fmt.Sprintf("Clients connected: %d", len(manager.clients))
				_, err := client.Write([]byte(numClients))
				if err != nil {
					fmt.Println("Couldn't send start message to client ", client)
				}
			}
			if len(manager.clients) == 3 {
				manager.startGame()
			}
		}

	}

}

func (manager *ClientManager) startGame() {
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Start Game")
	// Send message to clients to start game
	for _, client := range manager.clients {
		_, err := client.Write([]byte("Start Game"))
		if err != nil {
			fmt.Println("Couldn't send start message to client ", client)
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
				_, err := client.Write([]byte(message))
				if err != nil {
					fmt.Printf("Couldn't send message %+v to client %+v\n", message, client)
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

/*
	RECEIVE MESSAGES FROM CLIENTS
*/
func (manager *ClientManager) receiveMessages(client net.Conn) {
	for {
		message := make([]byte, 4096)
		length, err := client.Read(message)
		if err != nil {
			manager.disconnectClient <- client
			// fmt.Println("Error in socket connection,", err)
			client.Close()
			break
		}

		if length > 0 && manager.gameStarted == true {
			fmt.Println("RECEIVED: " + string(message))
			manager.receive <- []byte(message)
		}
	}
}
