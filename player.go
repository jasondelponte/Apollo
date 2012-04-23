package main

import (
	"log"
)

type PlayerCmd int

var (
	PlayerCmdGameSelectEntity = PlayerCmd(0)
)

type PlayerError struct {
	PlayerErrorString string
}

func (p *PlayerError) Error() string { return p.PlayerErrorString }

var (
	PlayerErrorDisconnected = &PlayerError{"Player's connection has been disconnected"}
)

type PlayerAction struct {
	World  *PlayerWorldAction
	Game   *PlayerGameAction
	Player *Player
}

type PlayerWorldAction struct {
}

type PlayerGameAction struct {
	Command  PlayerCmd
	EntityId EntityId
}

// Player object
type Player struct {
	id          uint64
	conn        Connection
	reader      chan MessageIn
	toPlayer    chan interface{}
	setGameCtrl chan *GamePlayerCtrl
	gameCtrl    GamePlayerCtrl
}

// Creates a new intance of the player object, and attaches the
// existing connection to the player.
func NewPlayer(id uint64, c Connection) *Player {
	p := &Player{
		id:   id,
		conn: c,
	}

	p.reader = make(chan MessageIn)
	p.toPlayer = make(chan interface{}, 10)
	p.setGameCtrl = make(chan *GamePlayerCtrl)
	p.conn.AttachReader(p.reader)

	return p
}

// Returns the player's id
func (p Player) GetId() uint64 {
	return p.id
}

// Terminates the player's connection
func (p *Player) Disconnect() {
	if p.conn != nil {
		p.conn.Close()
	}
	if p.toPlayer != nil {
		close(p.toPlayer)
	}
	if p.setGameCtrl != nil {
		close(p.setGameCtrl)
	}

	p.conn = nil
	p.toPlayer = nil
	p.setGameCtrl = nil
}

// Event handler for a player. Will process events as they are
// received from the player, world, or game
func (p *Player) Run(w *World) {
	defer func() { log.Println("Player ", p.id, " event loop terminating") }()
	for {
		select {
		case msg, ok := <-p.reader:
			if !ok {
				return
			}
			if msg.Act != nil {
				ctrl := GetPlayerActionFromMessage(msg, p)
				// forward the control onto the world or game
				if ctrl.Game != nil && p.gameCtrl != nil {
					p.gameCtrl <- ctrl
				}

				if ctrl.World != nil {
					w.playerAction <- ctrl
				}
			}

		case msg, ok := <-p.toPlayer:
			if !ok {
				return
			}
			if p.conn == nil {
				return
			}
			p.conn.Send(msg)

		case ctrl, ok := <-p.setGameCtrl:
			if !ok {
				return
			}
			if ctrl == nil {
				p.gameCtrl = nil
				continue
			}
			p.gameCtrl = *ctrl
		}
	}
}

// Pushes the message to the player asynchronously
func (p *Player) SendToPlayer(msg interface{}) error {
	if p.toPlayer == nil {
		return PlayerErrorDisconnected
	}

	p.toPlayer <- msg
	return nil
}

// Sets the channel a player should use to use to send controls
// to the the game at on.
func (p *Player) SetGameCtrl(ctrlChan *GamePlayerCtrl) error {
	if p.setGameCtrl == nil {
		return PlayerErrorDisconnected
	}

	p.setGameCtrl <- ctrlChan
	return nil
}
