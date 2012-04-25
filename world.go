package main

import (
	"log"
)

type WorldError struct {
	WorldErrorString string
}

func (w *WorldError) Error() string { return w.WorldErrorString }

var (
	WorldErrorPlayerNotRegistered = &WorldError{"Player is not registred"}
)

// The world object 
type World struct {
	nextGameId uint64
	players    map[*Player]*PlayerInstance
	games      []*Game

	register     chan *Player
	unregister   chan *Player
	playerAction chan *PlayerAction

	httpHndlr *HttpHandler
}

// Defines info about the player for this current instance 
// being connected to the world
type PlayerInstance struct {
	Game *Game
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewWorld(httpHndlr *HttpHandler) *World {
	w := &World{
		nextGameId: 0,
		players:    make(map[*Player]*PlayerInstance),
		games:      make([]*Game, 0, 10),

		register:     make(chan *Player),
		unregister:   make(chan *Player),
		playerAction: make(chan *PlayerAction),
		httpHndlr:    httpHndlr,
	}
	return w
}

// Event receiver to processing messages between the simulation and
// the players.  If players are connected to the game the simulation
// will be started, but as soon as the last player drops out the
// simulation will be terminated.
func (w *World) Run() {
	go w.httpHndlr.HandleHttpConnection(w)

	for {
		select {
		case p := <-w.register:
			log.Println("Registering player")
			err := w.registerPlayer(p)
			if err != nil {
				log.Println("Player failed to register: ", err)
				w.unregister <- p
			}

		case p := <-w.unregister:
			log.Println("Player unregistered")
			err := w.unregisterPlayer(p)
			if err != nil {
				log.Println("Failed to unregister player, ", err)
			}

		case ctrl := <-w.playerAction:
			info := w.players[ctrl.Player]
			if info == nil {
				ctrl.Player.Disconnect()
				continue
			}

			// TODO handle player world controls
		}
	}
}

// Registers the player with the world and randomly adds them to a game
// that is not full. If there are no available games a new one will be 
// created.
func (w *World) registerPlayer(p *Player) error {
	// TODO need some kind of logic for a player to specifiy the game type
	g := w.getAvailableGame(GameTypeMobileSmall)

	w.players[p] = &PlayerInstance{Game: g}

	// Kick off the player's event loop
	go p.Run(w)

	g.AddPlayer <- p

	return nil
}

// Returns a game object from the pool of available games
// If no available game exists, one will be created.
func (w *World) getAvailableGame(gameType *GameType) *Game {
	if len(w.games) == 0 {
		return w.addNewGame(gameType)
	}

	for _, g := range w.games[:] {
		if g != nil && g.gameType == gameType && !g.IsFull() {
			return g
		}
	}

	return w.addNewGame(gameType)
}

// Creates a new game and adds it to the list of games. The
// newly created game is also returned.
func (w *World) addNewGame(gameType *GameType) *Game {
	g := NewGame(w.nextGameId, gameType)
	w.nextGameId++
	w.games = append(w.games, g)
	go g.Run()

	return g
}

// Remvoes a game from the world's list of available games
func (w *World) removeGame(g *Game) {
	// TODO remove all players from a game, and terminate the game
}

// Removes a player from the world and all games they are connected to
func (w *World) unregisterPlayer(p *Player) error {
	var rtrn error = nil
	info := w.players[p]
	if info != nil {
		if info.Game != nil {
			info.Game.RmPlayer <- p
		}
		delete(w.players, p)
	} else {
		rtrn = WorldErrorPlayerNotRegistered
	}
	p.Disconnect()
	return rtrn
}
