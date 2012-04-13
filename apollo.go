package main

import (
	"flag"
	"log"
)

var addr = flag.String("a", ":8080", "Address and Port server is to run on")
var rootURLPath = flag.String("r", "/", "URL Path root of the webapp")

func main() {
	flag.Parse()

	defer func() { log.Println("Existing apollo") }()
	go game.Run()

	game.InitConnections(*addr, *rootURLPath)
}
