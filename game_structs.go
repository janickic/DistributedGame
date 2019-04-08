package main

import (
	"net"
	"sync"
	"time"
)

type Data interface{}

type Action int
type MessageType int

const (
	lock   Action = 0
	unlock Action = 1
	fill   Action = 2
)

const (
	dataError   MessageType = 0
	dataGame    MessageType = 1
	dataMove    MessageType = 2
	dataPlayer  MessageType = 3
	restartGame MessageType = 4
)

type Player struct {
	Id     int64 // For now id is idex in players array
	Ip     net.IP
	Colour int64
	Score  int64
}

type Cell struct {
	lock   sync.Mutex
	Locked bool
	Filled bool
	Owner  Player
}

type Game struct {
	Board        [][]Cell
	N            int
	MinFill      float32
	Players      [4]Player
	numOfPlayers int
	Active       bool
	id           int64
}

type Move struct {
	CellX     int
	CellY     int
	Action    Action
	Player    Player
	Timestamp time.Time
}

type Message struct {
	MsgType MessageType
	Body    Data
}

func (cell *Cell) Lock() {
	cell.lock.Lock()
}

func (cell *Cell) Unlock() {
	cell.lock.Unlock()
}

func (player *Player) IncreaseScore() {
	player.Score++
}
