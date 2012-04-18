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

type MsgBoardUpdates struct {
	BU []MsgBoardUpdateItem // Board Updates
}
type MsgBoardUpdateItem struct {
	T int         // Update Type
	E interface{} // Entity
}

// Game update message
type MsgGameUpdate struct {
	GU bool
	Ps []MsgPartPlayerInfo
	Es []MsgPartEntity
}
type MsgPartPlayerInfo struct {
	Id uint64
	N  string
	S  int
}
type MsgPartEntity struct {
	Id         uint64
	T          int
	S          int
	X, Y, W, H int
	R, G, B, A int
}

// Builds a game update message based on the list of the current elements.
// and the player game info.
func BuildGameUpdateMessage(players map[*Player]*GameScore, entities []*Entity) interface{} {
	msg := &MsgGameUpdate{
		GU: true,
		Ps: BuildGameUpdatePlayersMessage(players),
		Es: BuildGameUpdateEntitiesMessage(entities),
	}

	return msg
}

// Builds the player update message part
func BuildGameUpdatePlayersMessage(players map[*Player]*GameScore) []MsgPartPlayerInfo {
	if len(players) == 0 {
		return nil
	}

	ps := make([]MsgPartPlayerInfo, len(players))
	i := 0
	for p, s := range players {
		if p == nil || s == nil {
			log.Println("BuildGameUpdatePlayersMessage, invalid player")
			continue
		}

		ps[i].Id = p.GetId()
		ps[i].N = s.Name
		ps[i].S = s.Score
		i++
	}

	return ps
}

// Builds the entity update message part
func BuildGameUpdateEntitiesMessage(entities []*Entity) []MsgPartEntity {
	if len(entities) == 0 {
		return nil
	}

	es := make([]MsgPartEntity, len(entities))
	for i, e := range entities[:] {
		if e == nil {
			log.Println("BuildGameUpdateEntitiesMessage, invalid entity at ", i)
			continue
		}

		es[i].Id = e.id
		es[i].T = e.typ
		es[i].S = e.state
		es[i].X = e.pos.x
		es[i].Y = e.pos.y
		es[i].W = e.pos.width
		es[i].H = e.pos.height
		es[i].R = e.color.red
		es[i].G = e.color.green
		es[i].B = e.color.blue
		es[i].A = e.color.alpha

		// Todo entity specific items
	}

	return es
}
