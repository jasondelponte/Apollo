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

type EntityPos struct {
	x, y, width, height int
}

type EntityColor struct {
	red, green, blue, alpha int
}

type Entity struct {
	id        EntityId
	typ       EntityType
	state     EntityState
	ttl       time.Duration
	createdAt time.Time
	updatedAt time.Time
	pos       *EntityPos
	color     *EntityColor
}

// Create a new Entity as a Box
func NewBoxEntity(id EntityId, ttl time.Duration, pos *EntityPos, color *EntityColor) *Entity {
	return &Entity{
		id:    id,
		state: EntityStateAdded,
		typ:   EntityTypeBlock,
		ttl:   ttl,
		pos:   pos,
		color: color,
	}
}

// Returns the id for this entitiy
func (e Entity) GetId() EntityId {
	return e.id
}
