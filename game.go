package main

import (
	"fmt"
	"log"
	"time"
)

const (
	delayBetweenSimStep = (250 * time.Millisecond)
)

type GameState int
type GamePlayerState int

var (
	// Game state
	GameStateRunning = GameState(0)
	GameStatePaused  = GameState(1)
	GameStateStopped = GameState(2)
	// Game Player State 
	GamePlayerStateAdded   = GamePlayerState(0)
	GamePlayerStatePresent = GamePlayerState(1)
	GamePlayerStateUpdated = GamePlayerState(2)
	GamePlayerStateRemoved = GamePlayerState(3)
)

// Definition of the game object
type Game struct {
	id           uint64
	sim          *Simulation
	board        *Board
	state        GameState
	players      map[*Player]*GamePlayerInfo
	addPlayer    chan *Player
	removePlayer chan *Player
	playerAction chan *PlayerAction
}

type GamePlayerInfo struct {
	PlayerId uint64
	State    GamePlayerState
	Name     string
	Score    int
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewGame(id uint64) *Game {
	g := &Game{
		id:           id,
		state:        GameStateStopped,
		players:      make(map[*Player]*GamePlayerInfo),
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
	ticker := time.NewTicker(delayBetweenSimStep)
	defer func() {
		log.Println("Game ", g.id, " event loop terminating")
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			if g.state != GameStateRunning {
				continue
			}

			toA := g.sim.Step()
			if toA != nil {
				msg := MsgCreateGameUpdate()
				msg.AddEntityUpdates(toA)
				g.broadcastUpdate(msg)
			}

		case p := <-g.addPlayer:
			log.Printf("Adding player %d to game %d", p.GetId(), g.id)
			pInfo := &GamePlayerInfo{
				State:    GamePlayerStateAdded,
				PlayerId: p.GetId(),
				Score:    0,
				Name:     fmt.Sprintf("Player %d", p.GetId()),
			}
			if g.state == GameStateStopped {
				g.startGame()
			}

			toP := g.board.GetEntities()
			if toP != nil {
				msg := MsgCreateGameUpdate()
				infos := make([]*GamePlayerInfo, len(g.players))
				i := 0
				for _, info := range g.players {
					infos[i] = info
					i++
				}
				msg.AddPlayerGameInfos(infos)
				msg.AddEntityUpdates(toP)
				g.playerUpdate(p, msg)
			}
			// Don't add the new player to our list until it has already been updated.
			g.players[p] = pInfo

			toA := g.sim.PlayerJoined(p)
			if toA != nil {
				msg := MsgCreateGameUpdate()
				msg.AddPlayerGameInfo(pInfo, -1)
				msg.AddEntityUpdates(toA)
				g.broadcastUpdate(msg)
			}
			pInfo.State = GamePlayerStatePresent

		case p := <-g.removePlayer:
			log.Printf("Removing player %d from game %d", p.GetId(), g.id)
			if pInfo := g.players[p]; pInfo != nil {
				delete(g.players, p)

				pInfo.State = GamePlayerStateRemoved
				msg := MsgCreateGameUpdate()
				msg.AddPlayerGameInfo(pInfo, -1)
				g.broadcastUpdate(msg)
			}
			if len(g.players) == 0 {
				g.stopGame()
			}

		case ctrl := <-g.playerAction:
			var pInfo *GamePlayerInfo
			if pInfo = g.players[ctrl.Player]; pInfo == nil {
				// Ignore players we don't know about, should we disconnect them?
				continue
			}
			if ctrl.Game.Command == PlayerCmdGameRemoveEntity {
				e := g.board.RemoveEntityById(ctrl.Game.EntityId)
				if e != nil {
					pInfo.Score++
					pInfo.State = GamePlayerStateUpdated

					msg := MsgCreateGameUpdate()
					msg.AddPlayerGameInfo(pInfo, -1)
					msg.AddEntityUpdate(e, -1)
					g.broadcastUpdate(msg)
					pInfo.State = GamePlayerStatePresent
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
func (g Game) getState() GameState {
	return g.state
}
