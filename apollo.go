package main

import (
	"flag"
)

var addr = flag.String("a", "", "IP address the server is to run on")
var port = flag.Uint("p", 0, "Port address to run the server on")
var wsport = flag.Uint("wsport", 0, "Port the client will connect to the websockets on")
var rootURLPath = flag.String("r", "", "URL Path root of the webapp")
var servceStatic = flag.Bool("s", false, "Set if apollo should service up static content")
var wsConnType = flag.String("w", "gn", "Sets the websocket library to use, 'gn' for go.net, and 'gb' for garyburd/websocket")

func main() {
	flag.Parse()

	httpHndlr := &HttpHandler{
		Addr:        *addr,
		Port:        *port,
		WsPort:      *wsport,
		RootURLPath: *rootURLPath,
		ServeStatic: *servceStatic,
		WsConnType:  *wsConnType,
	}
	world := NewWorld(httpHndlr)

	world.Run()
}
