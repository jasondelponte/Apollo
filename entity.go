package main

import ()

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
	id    uint64
	typ   EntityType
	state EntityState
	pos   *EntityPos
	color *EntityColor
}

// Create a new Entity as a Box
func NewBoxEntity(id uint64, pos *EntityPos, color *EntityColor) *Entity {
	return &Entity{
		id:    id,
		state: EntityStateAdded,
		typ:   EntityTypeBlock,
		pos:   pos,
		color: color,
	}
}

// Returns the id for this entitiy
func (e Entity) GetId() uint64 {
	return e.id
}
