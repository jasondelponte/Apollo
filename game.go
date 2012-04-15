package main

import (
	"log"
)

type GameState struct {
	sigle byte
}

func (s *GameState) State() byte { return s.sigle }

var (
	GameStateRunning = &GameState{sigle: 0}
	GameStatePaused  = &GameState{sigle: 1}
	GameStateStopped = &GameState{sigle: 2}
)

// Definition of the game object
type Game struct {
	id           uint64
	sim          *Simulation
	board        *Board
	state        *GameState
	players      map[*Player]bool
	addPlayer    chan *Player
	removePlayer chan *Player
	playerCtrl   chan interface{}
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewGame(id uint64) *Game {
	g := &Game{
		id:           id,
		state:        GameStateStopped,
		players:      make(map[*Player]bool),
		addPlayer:    make(chan *Player),
		removePlayer: make(chan *Player),
		playerCtrl:   make(chan interface{}),
	}
	return g
}

func (g Game) GetId() uint64 {
	return g.id
}

// Event receiver to processing messages between the simulation and
// the players.  If players are connected to the game the simulation
// will be started, but as soon as the last player drops out the
// simulation will be terminated.
func (g *Game) Run() {
	log.Println("Game", g.id, "started")
	for {
		select {
		case <-g.addPlayer:
		case <-g.removePlayer:
		case <-g.playerCtrl:
			// TODO do soemthing with the incomming player control object
		}
	}
}

// Create the simulator, and start it running
func (g *Game) startGame() {
	// g.sim = NewSimulation()
	g.board = NewBoard()
	g.state = GameStateRunning
	go g.sim.Run()
}

// Terminate the simulator, and remove its instance
func (g *Game) stopGame() {
	close(g.sim.halt)
	g.sim = nil
	g.board = nil
}

// Returns the current state of the game
func (g Game) getState() *GameState {
	return g.state
}

// Returns if the game has reached its limit of players
func (g *Game) IsFull() bool {
	return false
}

// Adds a new player to the game.
func (g *Game) AddPlayer(p *Player) {
	log.Printf("Adding player %d to game %d", p.GetId(), g.id)
}
