package main

import ()

type Board struct {
	entities []*Entity
	players  []*Player
}

func NewBoard() *Board {
	return &Board{entities: make([]*Entity, 0, 10), players: make([]*Player, 0, 5)}
}

func (b *Board) AddPlayer(p *Player) {
	b.players = append(b.players, p)
}

func (b *Board) AddEntities(e []*Entity) {
	slice := b.entities
	orgLen := len(slice)
	newLen := orgLen + len(e)
	if newLen > cap(slice) { // reallocate
		// Allocate double what we have, for future growth.
		newSlice := make([]*Entity, 0, (newLen+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:newLen]
	copy(slice[orgLen:newLen], e)

	b.entities = slice
}

func (b *Board) AddEntity(e *Entity) {
	slice := b.entities
	l := len(slice)
	if l+1 > cap(slice) { // reallocate
		// Allocate double what we have, for future growth.
		newSlice := make([]*Entity, 0, (cap(slice)+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : l+1]
	slice[l+1] = e

	b.entities = slice
}
