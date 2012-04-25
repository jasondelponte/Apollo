package main

import (
	"time"
)

type EntityId uint64
type EntityType int
type EntityState int

var (
	// Entity Types
	EntityTypeBlock = EntityType(0)
	// Entity States
	EntityStateAdded    = EntityState(0)
	EntityStatePresent  = EntityState(1)
	EntityStateSelected = EntityState(2)
	EntityStateRemoved  = EntityState(3)
)

type Entity struct {
	id        EntityId
	typ       EntityType
	state     EntityState
	ttl       time.Duration
	createdAt time.Time
	updatedAt time.Time
	x         int
	y         int
	color     int
}

// Create a new Entity as a Box
func NewBoxEntity(id EntityId, ttl time.Duration, x, y, color int) *Entity {
	return &Entity{
		id:    id,
		state: EntityStateAdded,
		typ:   EntityTypeBlock,
		ttl:   ttl,
		x:     x,
		y:     y,
		color: color,
	}
}

// Returns the id for this entitiy
func (e Entity) GetId() EntityId {
	return e.id
}
