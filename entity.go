package main

import ()

const (
	ENTITY_TYPE_BLOCK    = 0
	ENTITY_STATE_REMOVED = 0
	ENTITY_STATE_ADDED   = 1
)

type EntityPos struct {
	x, y, width, height int
}

type EntityColor struct {
	red, green, blue, alpha int
}

type Entity struct {
	id    uint64
	typ   uint
	state int
	pos   *EntityPos
	color *EntityColor
}

// Create a new Entity as a Box
func NewBoxEntity(id uint64, pos *EntityPos, color *EntityColor) *Entity {
	return &Entity{
		id:    id,
		state: ENTITY_STATE_ADDED,
		typ:   ENTITY_TYPE_BLOCK,
		pos:   pos,
		color: color,
	}
}

// Returns the id for this entitiy
func (e Entity) GetId() uint64 {
	return e.id
}
