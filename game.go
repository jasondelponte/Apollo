package main

import (
	"fmt"
	"log"
	"time"
)

const (
	delayBetweenSimStep = (250 * time.Millisecond)
)

type GamePlayerCtrl chan *PlayerAction
type GameState int
type GamePlayerState int
type GameType struct {
	Rows, Cols, Players int
}

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
	// Game Types
	GameTypeMobileSmall = &GameType{Rows: 7, Cols: 5, Players: 5}
)

// Definition of the game object
type Game struct {
	id         uint64
	gameType   *GameType
	sim        *Simulation
	board      *Board
	state      GameState
	players    map[*Player]*GamePlayerInfo
	playerCtrl GamePlayerCtrl
	AddPlayer  chan *Player
	RmPlayer   chan *Player
	// Cache
	pInfoUpdates []*GamePlayerInfo
}

type GamePlayerInfo struct {
	PlayerId  PlayerId
	State     GamePlayerState
	Name      string
	Score     int
	SelcColor EntityColor
	Selected  []*Entity
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewGame(id uint64, gameType *GameType) *Game {
	g := &Game{
		id:         id,
		gameType:   gameType,
		state:      GameStateStopped,
		players:    make(map[*Player]*GamePlayerInfo),
		playerCtrl: make(GamePlayerCtrl),
		AddPlayer:  make(chan *Player),
		RmPlayer:   make(chan *Player),
		// Cache
		pInfoUpdates: make([]*GamePlayerInfo, 10),
	}
	return g
}

// Returns the game's id
func (g Game) GetId() uint64 {
	return g.id
}

// Returns if the game has reached its limit of players
func (g *Game) IsFull() bool {
	if g.gameType.Players <= len(g.players) {
		return true
	}

	return false
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
			g.simulate()

		case p := <-g.AddPlayer:
			log.Printf("Adding player %d to game %d", p.GetId(), g.id)
			g.addPlayer(p)

		case p := <-g.RmPlayer:
			log.Printf("Removing player %d from game %d", p.GetId(), g.id)
			g.removePlayer(p)

		case ctrl := <-g.playerCtrl:
			var pInfo *GamePlayerInfo
			if pInfo = g.players[ctrl.Player]; pInfo == nil {
				// Ignore players we don't know about, TODO should we disconnect them?
				continue
			}
			g.procPlayerCtrl(ctrl, pInfo)
		}
	}
}

func (g *Game) simulate() {
	toA := g.sim.Step()

	if toA != nil {
		g.pInfoUpdates = g.pInfoUpdates[0:0] // reinit the update cache

		// Searching over the changes for removed items someone has selected.
		for _, e := range toA {
			if e.state == EntityStateRemoved && e.Owner != nil {
				// Search through the players finding selections if there were any.
				pInfo := g.players[e.Owner]
				e.Owner = nil
				if pInfo == nil {
					log.Println("Removing entity owned by player that no longer exists")
					continue
				}

				// Update the player's score and remove the selected entities.
				for _, selc := range pInfo.Selected {
					if selc != nil {
						g.board.RemoveEntityById(selc.GetId())
						toA = append(toA, selc)
					}
				}
				if len(pInfo.Selected) > 1 {
					pInfo.Score += len(pInfo.Selected) - 1
				}
				pInfo.Selected = pInfo.Selected[0:0] // Clear this player's selection list
				pInfo.SelcColor = EntityNoColor

				g.pInfoUpdates = append(g.pInfoUpdates, pInfo)
			}
		}

		// Send the message
		msg := MsgCreateGameUpdate()
		msg.AddPlayerGameInfos(g.pInfoUpdates)
		msg.AddEntityUpdates(toA)
		g.broadcastUpdate(msg)
	}
}

// Adds a new player to the game, and starting the game if needed.
func (g *Game) addPlayer(p *Player) {
	pInfo := &GamePlayerInfo{
		State:     GamePlayerStateAdded,
		PlayerId:  p.GetId(),
		Score:     0,
		Name:      fmt.Sprintf("Player %d", p.GetId()),
		Selected:  make([]*Entity, 10),
		SelcColor: EntityNoColor,
	}
	pInfo.Selected = pInfo.Selected[0:0]
	if g.state != GameStateRunning {
		g.startGame()
	}

	// Update the current player with the current state of the game
	msg := MsgCreateGameUpdate()
	msg.AddGameType(g.gameType)
	// Let the new player know about the existing player list
	infos := make([]*GamePlayerInfo, len(g.players))
	i := 0
	for _, info := range g.players {
		infos[i] = info
		i++
	}
	msg.AddPlayerGameInfos(infos)
	// Get the entities and add them to the game if ther are any
	if toP := g.board.GetEntityArray(); toP != nil {
		msg.AddEntityUpdates(toP)
	}
	// send the message to the player
	g.playerUpdate(p, msg)

	g.players[p] = pInfo
	p.SetGameCtrl(&g.playerCtrl)

	// Let all players now about the new player
	msg = MsgCreateGameUpdate()
	msg.AddPlayerGameInfo(pInfo, -1)
	g.broadcastUpdate(msg)
}

// Removes the passed in player from the game, and stops the game
// if that is the last player to be removed.
func (g *Game) removePlayer(p *Player) {
	if pInfo := g.players[p]; pInfo != nil {
		delete(g.players, p)
		p.SetGameCtrl(nil)

		// Clear the ownership of these entities if there were any
		for _, e := range pInfo.Selected {
			if e == nil {
				continue
			}
			e.Owner = nil
			e.state = EntityStatePresent
		}

		// Let everyone else know the player left, and everything they had
		// is now unselected
		pInfo.State = GamePlayerStateRemoved
		msg := MsgCreateGameUpdate()
		msg.AddPlayerGameInfo(pInfo, -1)
		msg.AddEntityUpdates(pInfo.Selected)
		g.broadcastUpdate(msg)
	}
	if len(g.players) == 0 {
		g.stopGame()
	}
}

// Processes the player's control in relation to the game.
func (g *Game) procPlayerCtrl(ctrl *PlayerAction, pInfo *GamePlayerInfo) {
	if ctrl.Game.Command == PlayerCmdGameSelectEntity {
		// TODO do matching based on what the player selected previously
		pInfo.State = GamePlayerStateUpdated
		e := g.board.GetEntityById(ctrl.Game.EntityId)
		if e == nil { // the id wasn't found so ignore
			return
		}

		// unselect if already selected
		if e.state == EntityStateSelected {
			e.state = EntityStatePresent
			// remove tyhe old player's ref first
			oldPInfo := g.players[e.Owner]
			if oldPInfo != nil {
				for i, selc := range oldPInfo.Selected {
					if selc != nil && selc.GetId() == e.GetId() {
						oldPInfo.Selected[i] = nil
					}
				}
			}
			e.Owner = nil

		} else if pInfo.SelcColor == e.color || pInfo.SelcColor == EntityNoColor {
			e.state = EntityStateSelected
			e.Owner = ctrl.Player
			pInfo.SelcColor = e.color
			pInfo.Selected = append(pInfo.Selected, e)

		} else {
			e.state = EntityStatePresent
		}

		msg := MsgCreateGameUpdate()
		msg.AddPlayerGameInfo(pInfo, -1)
		msg.AddEntityUpdate(e, -1)
		g.broadcastUpdate(msg)

		pInfo.State = GamePlayerStatePresent
	}
}

// Processes an update from a player
func (g *Game) playerUpdate(p *Player, update interface{}) {
	err := p.SendToPlayer(update)
	if err != nil {
		g.RmPlayer <- p
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
	g.board = NewBoard(g.gameType.Rows, g.gameType.Cols)
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
