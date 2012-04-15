package main

import (
	"log"
)

type World struct {
	nextGameId uint64
	players    map[*Player]bool
	games      []*Game

	register   chan *Player
	unregister chan *Player
	playerCtrl chan interface{}

	httpHndlr *HttpHandler
}

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections.
func NewWorld(httpHndlr *HttpHandler) *World {
	w := &World{
		nextGameId: 0,
		players:    make(map[*Player]bool),
		games:      make([]*Game, 0, 10),

		register:   make(chan *Player),
		unregister: make(chan *Player),
		playerCtrl: make(chan interface{}),
		httpHndlr:  httpHndlr,
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
			if w.players[p] {
				delete(w.players, p)
			}
			p.Disconnect()

		case <-w.playerCtrl:
			// TODO do soemthing with the incomming player control object
		}
	}
}

// Registers the player with the world and randomly adds them to a game
// that is not full. If there are no available games a new one will be 
// created.
func (w *World) registerPlayer(p *Player) error {
	w.players[p] = true

	// Kick off the player's event loop
	go p.Run(w)

	g := w.getAvailableGame()
	g.AddPlayer(p)

	return nil
}

// Returns a game object from the pool of available games
// If no available game exists, one will be created.
func (w *World) getAvailableGame() *Game {
	if len(w.games) == 0 {
		return w.addNewGame()
	}

	for _, g := range w.games[:] {
		if g != nil && !g.IsFull() {
			return g
		}
	}

	return w.addNewGame()
}

// Creates a new game and adds it to the list of games. The
// newly created game is also returned.
func (w *World) addNewGame() *Game {
	g := NewGame(w.nextGameId)
	w.nextGameId += 1
	w.games = append(w.games, g)
	go g.Run()

	return g
}

// Remvoes a game from the world's list of available games
func (w *World) removeGame(g *Game) {
}
