package main

import (
	"math/rand"
	"time"
)

type Simulation struct {
	nextEntityId EntityId
	board        *Board

	// persistant temp storage
	toRmList     []*Entity
	toUpdateList []*Entity
	lastAdd      time.Time
}

// Create a new instance of the simulator
func NewSimulation(b *Board) *Simulation {
	return &Simulation{
		board:        b,
		toRmList:     make([]*Entity, 5),
		toUpdateList: make([]*Entity, 10),
	}
}

// Incremental update to the board, returns the entities updated
func (s *Simulation) Step() []*Entity {
	timeNow := time.Now().UTC()

	// get empty slices from the perisitant arrays
	toRmList := s.toRmList[0:0]
	toRemove := 0
	toUpdateList := s.toUpdateList[0:0]
	toUpdate := 0

	es := s.board.GetEntities()
	for _, e := range es {
		if timeNow.Sub(e.updatedAt) >= e.ttl {
			toRmList = append(toRmList, e)
			toRemove++

		}
		// TODO not sure what to do with just updated yet.
	}

	// Remove the entities which are no longer valid.
	for _, e := range toRmList {
		s.board.RemoveEntityById(e.id)
		// Add the items to be removed to the update list to be returned
		toUpdateList = append(toUpdateList, e)
		toUpdate++
	}

	// Adds new entities, and update the list
	toUpdateList = s.addNew(toUpdateList)

	// Update the persistant objects
	s.toRmList = toRmList
	s.toUpdateList = toUpdateList

	// return the list of updates
	if len(toUpdateList) > 0 {
		return toUpdateList
	}

	return nil
}

// Adds new entities and updates the list as needed
func (s *Simulation) addNew(list []*Entity) []*Entity {
	c := rand.Intn(3)
	for i := 0; i < c; i++ {
		if rand.Intn(10) >= 8 {
			list = append(list, s.addRandomBlock())
		}
	}

	return list
}

// Creates a new random block and adds it to the board
// The reference to the block created will be returned
func (s *Simulation) addRandomBlock() *Entity {
	e := NewBoxEntity(s.nextEntityId,
		time.Duration(rand.Intn(2500)+2500)*time.Millisecond,
		EntityPos{x: rand.Intn(320), y: rand.Intn(480), width: 30, height: 30},
		rand.Intn(5),
	)
	s.nextEntityId++
	s.board.AddEntity(e)
	return e
}
