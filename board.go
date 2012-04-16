package main

import (
	"log"
)

type Board struct {
	entities map[*Entity]bool
}

// Creates and initialies a new board
func NewBoard() *Board {
	return &Board{entities: make(map[*Entity]bool)}
}

// Adds a single entity to the board
func (b *Board) AddEntity(e *Entity) {
	log.Println("Adding new entitiy,", e.GetId())
	b.entities[e] = true
}

// Returns an array of the current entities
func (b *Board) GetEntities() []*Entity {
	if len(b.entities) == 0 {
		return nil
	}

	entities := make([]*Entity, len(b.entities))[:]
	var count int = 0

	for e, _ := range b.entities {
		entities[count] = e
		count++
	}

	return entities
}

// Removes an entity from the board. Returns true if the 
// entity was found and removed, otherwise false is returned.
func (b *Board) RemoveEntityById(id uint64) *Entity {
	var found *Entity = nil
	for e, _ := range b.entities {
		if e.GetId() == id {
			found = e
			break
		}
	}

	if found != nil {
		log.Println("removeing entitiy,", id)
		delete(b.entities, found)
		return found
	}

	log.Println("Could not find and remove entitity,", id)
	return nil
}
