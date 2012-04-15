package main

import (
	"math/rand"
)

type Simulation struct {
	board *Board
}

// Create a new instance of the simulator
func NewSimulation(b *Board) *Simulation {
	return &Simulation{board: b}
}

func (s *Simulation) Step() *MsgBoardUpdates {
	updates := &MsgBoardUpdates{
		BU: make([]*MsgBoardUpdateItem, 0, 1),
	}

	var update MsgBoardUpdateItem
	update.T = UPDATE_TYPE_ADD
	update.E = &MsgBlock{
		T: ENTITY_TYPE_BLOCK,
		X: rand.Intn(285),
		Y: rand.Intn(285),
		R: rand.Intn(255),
		G: rand.Intn(255),
		B: rand.Intn(255),
		A: rand.Intn(100),
		W: 30,
		H: 30,
	}

	updates.BU = append(updates.BU, &update)

	return updates
}

func (s *Simulation) UpdateBoard(action interface{}) {

}

func (s Simulation) GetCurrentBoard() {

}
