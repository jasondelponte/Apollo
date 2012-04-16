package main

import (
	"math/rand"
)

type Simulation struct {
	nextEntityId uint64
	board        *Board
}

// Create a new instance of the simulator
func NewSimulation(b *Board) *Simulation {
	return &Simulation{board: b}
}

// Incremental update to the board, returns the entities updated
func (s *Simulation) Step() []*Entity {
	randLen := rand.Intn(5)
	entities := make([]*Entity, 0, randLen)
	for i := 0; i < randLen; i++ {
		r := rand.Intn(9)
		if r > 5 && r <= 9 {
			entities = append(entities, s.addRandomBlock())

		} else if r < 3 && r >= 0 {
			e := s.board.GetRandomEntity()
			if e == nil {
				continue
			}
			entities = append(entities, s.board.RemoveEntityById(e.GetId()))
		}
	}

	if len(entities) > 0 {
		return entities
	}

	return nil
}

// Returns a list of entities for the current board
func (s *Simulation) GetCurrentBoard() []*Entity {
	return s.board.GetEntities()
}

// Updates the board reflecting a new player has joined.
func (s *Simulation) PlayerJoined(p *Player) []*Entity {
	e := s.addRandomBlock()
	entities := make([]*Entity, 1)
	entities[0] = e
	return entities
}

// Creates a new random block and adds it to the board
// The reference to the block created will be returned
func (s *Simulation) addRandomBlock() *Entity {
	e := NewBoxEntity(s.nextEntityId,
		&EntityPos{
			x: rand.Intn(285), y: rand.Intn(285),
			width: 30, height: 30,
		},
		&EntityColor{
			red: rand.Intn(255), green: rand.Intn(255), blue: rand.Intn(255),
			alpha: rand.Intn(50) + 50,
		},
	)
	s.nextEntityId++

	s.board.AddEntity(e)
	return e
}
