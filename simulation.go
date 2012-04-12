package main

import (
	"math/rand"
	"time"
)

type Simulation struct {
	halt      chan bool
	simUpdate chan interface{}
}

// Create a new instance of the simulator
func NewSimulation(update chan interface{}) *Simulation {
	return &Simulation{halt: make(chan bool), simUpdate: update}
}

// Simulator's processor loop. It waits for the either the halt channel
// to be closed or the timer to expire before doing any work. If the timer
// expires a random block will be created and sent to the the game
func (s *Simulation) Run() {
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			update := make(map[string]*MsgBlock, 1)

			update["Block"] = &MsgBlock{
				X: rand.Intn(285),
				Y: rand.Intn(285),
				R: rand.Intn(255),
				G: rand.Intn(255),
				B: rand.Intn(255),
				A: rand.Intn(100),
				W: 30,
				H: 30,
			}
			s.simUpdate <- update
		case _, ok := <-s.halt:
			if !ok {
				return
			}
		}

	}
}
