package main

import (
	"log"
)

// Player object
type Player struct {
	id       uint64
	conn     Connection
	reader   chan MessageIn
	toPlayer chan interface{}
}

// Creates a new intance of the player object, and attaches the
// existing connection to the player.
func NewPlayer(id uint64, c Connection) *Player {
	p := &Player{
		id:   id,
		conn: c,
	}

	p.reader = make(chan MessageIn)
	p.toPlayer = make(chan interface{})
	p.conn.AttachReader(p.reader)

	return p
}

func (p Player) GetId() uint64 {
	return p.id
}

// Terminates the player's connection
func (p *Player) Disconnect() {
	if p.conn == nil {
		return
	}

	p.conn.Close()
}

// Event handler for a player. Will process events as they are
// received from the player, world, or game
func (p *Player) Run(w *World) {
	for {
		select {
		case msg, ok := <-p.reader:
			if !ok {
				return
			}
			log.Println("Received player message: ", msg)
			// TODO Probably want to do something with this message
			// instead of just forwarding it to the game.
			// game.playerCtrl <- msg
		case msg, ok := <-p.toPlayer:
			if !ok {
				return
			}
			if p.conn == nil {
				return
			}
			p.conn.Send(msg)
		}
	}
}

// Pushes the message to the player asynchronously
// TODO this is actually blocking until the player's event loop
// can read from the channel, so the toPlayer chan should be
// buffered with a go routine
func (p *Player) SendToPlayer(msg interface{}) {
	p.toPlayer <- msg
}
