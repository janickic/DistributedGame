package main

import (
	"bufio"
	"encoding/gob"
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

var cur_game = Game{}
var my_player = Player{}

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
		gob_decoder := gob.NewDecoder(client.socket)
		err := gob_decoder.Decode(message)
		if err != nil {
			fmt.Println("decoding error: ", err)
		}
		
		if message.Msg_type == data_game {
			cur_game = message.Body.(Game)
			fmt.Println("Received Game")
		} else if message.Msg_type == data_player {
			my_player = message.Body.(Player)
			fmt.Println("I am player", my_player.Id)
		}
	}
}

func (client *Client) chanReceive() {
	for {
		data := <-client.data
		fmt.Println("RECEIVED: " + string(data))
	}
}
