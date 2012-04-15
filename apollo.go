package main

import (
	"flag"
)

var addr = flag.String("a", "", "IP address the server is to run on")
var port = flag.Uint("p", 8080, "Port address to run the server on")
var rootURLPath = flag.String("r", "", "URL Path root of the webapp")
var servceStatic = flag.Bool("s", false, "Set if apollo should service up static content")

func main() {
	flag.Parse()

	httpHndlr := &HttpHandler{Addr: *addr, Port: *port, RootURLPath: *rootURLPath, ServeStatic: *servceStatic}
	world := NewWorld(httpHndlr)

	world.Run()
}
