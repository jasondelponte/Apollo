package main

import (
	"flag"
	"log"
)

var addr = flag.String("addr", ":8080", "Address and Port server is to run on")

func main() {
	flag.Parse()

	defer func() { log.Println("Existing apollo") }()
	go game.Run()

	game.InitConnections(*addr)
}
