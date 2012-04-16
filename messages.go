package main

import (
	"log"
)

// Incoming message container which contains all possible 
// input message types
type MessageIn struct {
	ReqId string
	Act   *MsgPlayerAction
}

type MsgPlayerAction struct {
	W *MsgPartActionWorld
	G *MsgPartActionGame
}

type MsgPartActionWorld struct {
}

type MsgPartActionGame struct {
	C int
	E uint64
}

// Builds the player control object from the message
func GetPlayerActionFromMessage(msg MessageIn, p *Player) *PlayerAction {
	if msg.Act == nil {
		return nil
	}

	action := &PlayerAction{Player: p}
	if msg.Act.W != nil {
		action.World = &PlayerWorldAction{}

	}

	if msg.Act.G != nil {
		action.Game = &PlayerGameAction{
			Command:  msg.Act.G.C,
			EntityId: msg.Act.G.E,
		}
	}

	return action
}

// Response messages
type MsgBoardCurrent struct {
	E []interface{} // Entities
}
type MsgBoardUpdates struct {
	BU []MsgBoardUpdateItem // Board Updates
}
type MsgBoardUpdateItem struct {
	T int         // Update Type
	E interface{} // Entity
}

// Message part for a block
type MsgBlock struct {
	ID uint64
	T  int
	S  int
	X  int
	Y  int
	R  int
	G  int
	B  int
	A  int
	W  int
	H  int
}

// Creates the board updated message. contains a list of updates
// of the board's entities
func BuildBoardUpdateMessage(updates []*Entity) interface{} {
	msg := &MsgBoardUpdates{
		BU: make([]MsgBoardUpdateItem, len(updates)),
	}

	for i, e := range updates[:] {
		if e == nil {
			log.Println("BuildBoardUpdateMessage, invalid entity at ", i)
			continue
		}
		if e.typ == ENTITY_TYPE_BLOCK {
			buildMsgBlockFromEntity(e, &msg.BU[i])
		}
	}

	return msg
}

// Builds the message for an entity being remoed
func BuildBoardUpdateMessageSingle(e *Entity) interface{} {
	msg := &MsgBoardUpdates{
		BU: make([]MsgBoardUpdateItem, 1),
	}

	if e.typ == ENTITY_TYPE_BLOCK {
		buildMsgBlockFromEntity(e, &msg.BU[0])
	}

	return msg
}

// Builds the message item for a block from an entity
func buildMsgBlockFromEntity(entity *Entity, updateItem *MsgBoardUpdateItem) {
	updateItem.T = entity.state
	updateItem.E = &MsgBlock{
		ID: entity.id,
		T:  ENTITY_TYPE_BLOCK,
		S:  entity.state,
		X:  entity.pos.x,
		Y:  entity.pos.y,
		W:  entity.pos.width,
		H:  entity.pos.height,
		R:  entity.color.red,
		G:  entity.color.green,
		B:  entity.color.blue,
		A:  entity.color.alpha,
	}
}
