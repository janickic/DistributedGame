package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Client provides socket info and data for single client
type Client struct {
	socket net.Conn
	data   chan []byte
}

func startClientMode(ip string) {
	// TODO: check number of connected clients
	// Once 3 have been connected, start a game

	fmt.Println("Starting client...")
	ip = fmt.Sprintf("%s:12345", ip)
	connection, error := net.Dial("tcp", ip)
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}

	go client.receive()
	/*
		SEND MESSAGES TO SERVER
	*/
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		// Send data to server
		connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
}

/*
	RECEIVE MESSAGES TO SERVER
*/
func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}
