package main

import (
	"fmt"
	"log"
	"time"
)

const (
	delayBetweenSimStep = (250 * time.Millisecond)
)

type GameState struct {
	s byte
}

func (s *GameState) State() byte { return s.s }

var (
	GameStateRunning = &GameState{s: 0}
	GameStatePaused  = &GameState{s: 1}
	GameStateStopped = &GameState{s: 2}
)

// Definition of the game object
type Game struct {
	id           uint64
	sim          *Simulation
	board        *Board
	state        *GameState
	players      map[*Player]*GameScore
	addPlayer    chan *Player
	removePlayer chan *Player
	playerAction chan *PlayerAction
}

type GameScore struct {
	Name  string
	Score int
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewGame(id uint64) *Game {
	g := &Game{
		id:           id,
		state:        GameStateStopped,
		players:      make(map[*Player]*GameScore),
		addPlayer:    make(chan *Player),
		removePlayer: make(chan *Player),
		playerAction: make(chan *PlayerAction),
	}
	return g
}

// Returns the game's id
func (g Game) GetId() uint64 {
	return g.id
}

// Returns if the game has reached its limit of players
func (g *Game) IsFull() bool {
	return false
}

// Signals the game to add a new player to the game
func (g *Game) AddPlayer(p *Player) {
	g.addPlayer <- p
}

// Signals the game to remove a player from the game
func (g *Game) RemovePlayer(p *Player) {
	g.removePlayer <- p
}

// Event receiver to processing messages between the simulation and
// the players.  If players are connected to the game the simulation
// will be started, but as soon as the last player drops out the
// simulation will be terminated.
func (g *Game) Run() {
	defer func() { log.Println("Game ", g.id, " event loop terminating") }()
	ticker := time.NewTicker(delayBetweenSimStep)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if g.state != GameStateRunning {
				continue
			}

			toA := g.sim.Step()
			if toA != nil {
				g.broadcastUpdate(BuildGameUpdateMessage(g.players, toA))
			}

		case p := <-g.addPlayer:
			log.Printf("Adding player %d to game %d", p.GetId(), g.id)
			g.players[p] = &GameScore{Score: 0, Name: fmt.Sprintf("Player %d", p.GetId())}
			if g.state == GameStateStopped {
				g.startGame()
			}

			toP := g.board.GetEntities()
			if toP != nil {
				g.playerUpdate(p, BuildGameUpdateMessage(g.players, toP))
			}

			toA := g.sim.PlayerJoined(p)
			if toA != nil {
				g.broadcastUpdate(BuildGameUpdateMessage(g.players, toA))
			}

		case p := <-g.removePlayer:
			log.Printf("Removing player %d from game %d", p.GetId(), g.id)
			if g.players[p] != nil {
				delete(g.players, p)
			}
			if len(g.players) == 0 {
				g.stopGame()
			}

		case ctrl := <-g.playerAction:
			var gs *GameScore
			if gs = g.players[ctrl.Player]; gs == nil {
				// Ignore players we don't know about, should we disconnect them?
				continue
			}
			if ctrl.Game.Command == PLAYER_CMD_GAME_REMOVE_ENTITY {
				e := g.board.RemoveEntityById(ctrl.Game.EntityId)
				if e != nil {
					gs.Score++
					es := make([]*Entity, 1)
					es[0] = e
					g.broadcastUpdate(BuildGameUpdateMessage(g.players, es))
				}
			}
		}
	}
}

// Processes an update from a player
func (g *Game) playerUpdate(p *Player, update interface{}) {
	err := p.SendToPlayer(update)
	if err != nil {
		g.RemovePlayer(p)
	}
}

// Sends out an update to all players
func (g *Game) broadcastUpdate(update interface{}) {
	for p, _ := range g.players {
		g.playerUpdate(p, update)
	}
}

// Create the simulator, and start it running
func (g *Game) startGame() {
	g.board = NewBoard()
	g.sim = NewSimulation(g.board)
	g.state = GameStateRunning
}

// Terminate the simulator, and remove its instance
func (g *Game) stopGame() {
	g.state = GameStateStopped
	g.sim = nil
	g.board = nil
}

// Returns the current state of the game
func (g Game) getState() *GameState {
	return g.state
}
