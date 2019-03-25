package main

import (
	"net"
	"sync"
	"time"
)

type Data interface{}

type Action int
type MessageType int

const(
	lock Action = 0
	unlock Action = 1
	fill Action = 2 
)

const(
	data_player MessageType = 0
	data_game MessageType = 1
	data_move MessageType = 2
)

type Player struct{
	Id int64 // For now id is idex in players array
	Ip net.IP
	Colour int64
	Score int64
}

type Cell struct{
	sync.Mutex
	Locked bool
	Filled bool
	Owner Player
}

type Game struct{
	Board [][]Cell
	N int
	Min_fill float32
	Players [4]Player
	Active bool
}

type Move struct{
	Cell_x int
	Cell_y int
	Action Action
	Player Player
	Timestamp time.Time
}

type Message struct{
	Msg_type MessageType
	Body Data
}