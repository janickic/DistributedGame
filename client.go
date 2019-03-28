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

var game = new(Game)

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
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			client.data <- message
		}
	}
}

func (client *Client) chanReceive() {
	for {
		data := <-client.data
		fmt.Println("RECEIVED: " + string(data))
	}
}
