package main

import ()

const (
	ENTITY_TYPE_BLOCK = 0
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
	pos   *EntityPos
	color *EntityColor
}

// Create a new Entity as a Box
func NewBoxEntity(id uint64, pos *EntityPos, color *EntityColor) *Entity {
	return &Entity{
		id:    id,
		typ:   ENTITY_TYPE_BLOCK,
		pos:   pos,
		color: color,
	}
}

// Returns the id for this entitiy
func (e Entity) GetId() uint64 {
	return e.id
}
