package main

import (
	"github.com/garyburd/go-websocket/websocket"
	"html/template"
	"log"
	"net/http"
)

type Game struct {
	sim        *Simulation
	players    map[*Player]bool
	register   chan *Player
	unregister chan *Player
	playerCtrl chan interface{}
	simUpdate  chan interface{}
}

var homeTempl = template.Must(template.ParseFiles("templates/home.html"))

// Initalization of the game object.game  It s being done in the package's
// global scope so the network event handler will have access to it when
// receiving new player connections. TODO figure out how to remove it from global
var game = &Game{
	players:    make(map[*Player]bool),
	register:   make(chan *Player),
	unregister: make(chan *Player),
	playerCtrl: make(chan interface{}),
	simUpdate:  make(chan interface{}),
}

// Creates and setups the the network event listers for html
// and websocket interfaces
func (g *Game) InitConnections(addr string) {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWsConn)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Event receiver to processing messages between the simulation and
// the players.  If players are connected to the game the simulation
// will be started, but as soon as the last player drops out the
// simulation will be terminated.
func (g *Game) Run() {
	for {
		select {
		case p := <-g.register:
			log.Println("New Player registered")
			g.players[p] = true

			// Create the sim if this is the first user connected
			if g.sim == nil {
				g.startSim()
			}

		case p := <-g.unregister:
			log.Println("Player unregistered")
			delete(g.players, p)
			p.Disconnect()

			// Disable the sim if there are no more players
			if len(g.players) == 0 {
				g.stopSim()
			}
		case <-g.playerCtrl:
			// TODO do soemthing with the incomming player control object

		case update := <-g.simUpdate:
			for p, _ := range g.players {
				p.UpdateBoard(update)
			}
		}

	}
}

// Create the simulator, and start it running
func (g *Game) startSim() {
	g.sim = NewSimulation(g.simUpdate)
	go g.sim.Run()
}

// Terminate the simulator, and remove its instance
func (g *Game) stopSim() {
	close(g.sim.halt)
	g.sim = nil
}

// Network event handler for HTTP trafic. Serves up the 
// home.html file which will allow connection to the websocket
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method nod allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

// handles webocket requests from the client.
func serveWsConn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := websocket.Upgrade(w, r.Header, "", 1024, 1024)
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad request", 400)
		return
	}

	conn := &WsConn{send: make(chan []byte, 256), ws: ws}
	player := NewPlayer(conn)

	defer func() { game.unregister <- player }()
	go conn.WritePump()
	go player.Run(game)

	// Read pump will hold the connection open until we are finished with it.
	conn.ReadPump()
}
