package main

import (
	"math/rand"
	"time"
)

const (
	MinTimeBetweenAdds = (1000 * time.Millisecond)
)

type Simulation struct {
	nextEntityId EntityId
	board        *Board
	lastAddedOn  time.Time

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
	// get empty slices from the perisitant arrays
	toRmList := s.toRmList[0:0]
	toUpdateList := s.toUpdateList[0:0]

	es := s.board.GetEntities()
	for _, e := range es {
		if time.Since(e.updatedAt) >= e.ttl {
			toRmList = append(toRmList, e)
		}
		// TODO not sure what to do with just updated yet.
	}

	// Remove the entities which are no longer valid.
	for _, e := range toRmList {
		s.board.RemoveEntityById(e.id)
		// Add the items to be removed to the update list to be returned
		toUpdateList = append(toUpdateList, e)
	}

	// Adds new entities, and update the list
	if time.Since(s.lastAddedOn) >= MinTimeBetweenAdds {
		toUpdateList = s.addNew(toUpdateList)
		s.lastAddedOn = time.Now().UTC()
	}

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
	c := rand.Intn(5)
	for i := 0; i < c; i++ {
		if e := s.addRandomBlock(); e != nil {
			list = append(list, e)
		}
	}

	return list
}

// Creates a new random block and adds it to the board
// The reference to the block created will be returned
func (s *Simulation) addRandomBlock() *Entity {
	x := rand.Intn(s.board.Cols)
	y := rand.Intn(s.board.Rows)
	if s.board.EntityAtPos(x, y) {
		// don't create duplicate blocks at the same point
		return nil
	}

	e := NewBoxEntity(s.nextEntityId,
		time.Duration(7000)*time.Millisecond,
		x, y,
		EntityColor(rand.Intn(5)),
	)
	s.nextEntityId++
	s.board.AddEntity(e)
	return e
}
