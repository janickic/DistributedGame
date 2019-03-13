package main

import (
	"fmt"
	"net"
	"sync"
)

// ClientManager holds available clients
type ClientManager struct {
	clients []net.Conn
	lock    sync.RWMutex
	receive chan []byte
}

func startServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", ":12345")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients: make([]net.Conn, 0, 4),
		receive: make(chan []byte),
	}

	// Wait for 3 connections
	manager.registerConnections(listener)
	fmt.Println("Start game")
	// Send message to clients to start game
	for _, client := range manager.clients {
		_, err := client.Write([]byte("Start Game"))
		if err != nil {
			fmt.Println("Couldn't send start message to client ", client)
		}
		go manager.receiveMessages(client)
	}

	for {
		// Receive message from channel
		message := <-manager.receive
		// Send message to all other clients
		for _, client := range manager.clients {
			_, err := client.Write([]byte(message))
			if err != nil {
				fmt.Printf("Couldn't send message %+v to client %+v\n", message, client)
			}
		}
	}

}

func (manager *ClientManager) registerConnections(listener net.Listener) {
	for {
		connection, err := listener.Accept()
		fmt.Println("Client connected, ", len(manager.clients)+1)
		if err != nil {
			fmt.Println(err)
		}
		manager.lock.Lock()
		manager.clients = append(manager.clients, connection)
		if len(manager.clients) == 3 {
			manager.lock.Unlock()
			return
		}
		manager.lock.Unlock()
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
			fmt.Println("Error in socket connection,", err)
			client.Close()
			break
		}

		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
			manager.receive <- []byte(message)
		}
	}
}
