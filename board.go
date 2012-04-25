package main

import (
	"log"
	"math/rand"
	"time"
)

type EntityMap map[EntityId]*Entity

type Board struct {
	Rows, Cols int // Defines the number for rows and columns a board is
	entities   EntityMap
}

// Creates and initialies a new board
func NewBoard(rows, cols int) *Board {
	return &Board{
		Rows:     rows,
		Cols:     cols,
		entities: make(EntityMap),
	}
}

// Adds a single entity to the board
func (b *Board) AddEntity(e *Entity) {
	log.Println("Adding new entitiy,", e.GetId())
	e.createdAt = time.Now().UTC()
	e.updatedAt = e.createdAt
	b.entities[e.id] = e
}

func (b *Board) EntityAtPos(x, y int) bool {
	for _, e := range b.entities {
		if e.x == x && e.y == y {
			return true
		}
	}
	return false
}

// Returns the entity map on the board
func (b *Board) GetEntities() EntityMap {
	return b.entities
}

// Returns an array of the current entities
func (b *Board) GetEntityArray() []*Entity {
	if len(b.entities) == 0 {
		return nil
	}

	entities := make([]*Entity, len(b.entities))[:]

	i := 0
	for _, e := range b.entities {
		entities[i] = e
		i++
	}

	return entities
}

// Returns an entity at the id specified
func (b *Board) GetEntityById(id EntityId) *Entity {
	return b.entities[id]
}

// Removes an entity from the board. Returns true if the 
// entity was found and removed, otherwise false is returned.
func (b *Board) RemoveEntityById(id EntityId) *Entity {
	e := b.entities[id]

	if e != nil {
		log.Println("Removing entity,", e.id)
		e.state = EntityStateRemoved
		delete(b.entities, e.id)
		return e
	}

	log.Println("Could not find and remove entitity,", id)
	return nil
}

// Returns an random entity
func (b *Board) GetRandomEntity() *Entity {
	i := 0
	r := 0
	n := len(b.entities)
	if n <= 0 {
		return nil
	} else if n == 1 {
		r = n
	} else {
		r = rand.Intn(n - 1)
	}

	for _, e := range b.entities {
		if i == r {
			return e
		}
		i++
	}

	return nil
}
