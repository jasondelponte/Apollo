package main

import (
	"flag"
	"log"
)

var addr = flag.String("a", "", "Address server is to run on")
var port = flag.Uint("p", 8080, "Port address to run the server on")
var rootURLPath = flag.String("r", "/", "URL Path root of the webapp")

func main() {
	flag.Parse()

	defer func() { log.Println("Existing apollo") }()
	go game.Run()

	game.InitConnections()
}
