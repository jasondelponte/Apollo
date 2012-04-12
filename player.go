package main

import ()

type Player struct {
	id     uint64
	conn   Connection
	reader chan MessageIn
}

// Creates a new intance of the player object, and attaches the
// existing connection to the player.
func NewPlayer(c Connection) *Player {
	p := &Player{conn: c}

	p.reader = make(chan MessageIn)
	p.conn.AttachReader(p.reader)

	return p
}

// Terminates the player's connection
func (p *Player) Disconnect() {
	p.conn.Close()
}

// Event handler for processing events received from the player
// Any message received from the player will be forwarded to the game.
func (p *Player) Run(game *Game) {

	game.register <- p

	for {
		select {
		case msg, ok := <-p.reader:
			if !ok {
				return
			}
			// TODO Probably want to do something with this message
			// instead of just forwarding it to the game.
			game.playerCtrl <- msg
		}
	}
}

// Allows the game to update the player with the the latest simulations update.
func (p *Player) UpdateBoard(update interface{}) {
	p.conn.Send(update)
}
