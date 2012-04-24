package main

import (
	"time"
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
			Command:  PlayerCmd(msg.Act.G.C),
			EntityId: EntityId(msg.Act.G.E),
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
	Id uint64 // Id
	St int    // State of the player
	N  string // Name
	Sc int    // Score
}
type MsgPartEntity struct {
	Id         uint64 // entity ID
	T          int    // Type of the entity
	St         int    // State of the entity
	X, Y, W, H int    // position
	C          int    // color
	Ttl        int64  // Time to live, in miliseconds
	CAt, UAt   int64  // Create and Update Time
}

func MsgCreateGameUpdate() *MsgGameUpdate {
	return &MsgGameUpdate{GU: true}
}

// Grows the player game info list of game update if needed
// to fit the extra legthn needed.
func (m *MsgGameUpdate) growPlayerGameInfosToFit(addLen int) {
	if m.Ps != nil {
		m.Ps = make([]MsgPartPlayerInfo, addLen)
	} else if len(m.Ps)+addLen > cap(m.Ps) {
		newPs := make([]MsgPartPlayerInfo, len(m.Ps)+addLen)
		copy(newPs, m.Ps)
		m.Ps = newPs
	}
}

// Grows the entity update list of game update if needed
// to fit the extra legthn needed.
func (m *MsgGameUpdate) growEntityUpdatesToFit(addLen int) {
	if m.Es != nil {
		m.Es = make([]MsgPartEntity, addLen)
	} else if len(m.Es)+addLen > cap(m.Es) {
		newEs := make([]MsgPartEntity, len(m.Es)+addLen)
		copy(newEs, m.Es)
		m.Es = newEs
	}
}

// Adds a list of player game infos to the update message. This
// will auto grow the message as needed.
func (m *MsgGameUpdate) AddPlayerGameInfos(infos []*GamePlayerInfo) {
	m.growPlayerGameInfosToFit(len(infos))

	for i, info := range infos {
		m.AddPlayerGameInfo(info, i)
	}
}

// Adds a single player game info to the update message. This
// Will auto grow the message as needed. The second parameter is
// the the index to insert the info into the message. if it is -1
// the message will be appended to the end of the message. If the 
// idx is provided a check for message size is nto made.
func (m *MsgGameUpdate) AddPlayerGameInfo(info *GamePlayerInfo, idx int) {
	i := idx
	if i == -1 {
		i = len(m.Ps)
		m.growPlayerGameInfosToFit(1)
	}

	m.Ps[i].Id = info.PlayerId
	m.Ps[i].St = int(info.State)
	m.Ps[i].N = info.Name
	m.Ps[i].Sc = info.Score
}

// Adds a list of entities to the update message. This will auto
// grow the message as needed
func (m *MsgGameUpdate) AddEntityUpdates(entities []*Entity) {
	m.growEntityUpdatesToFit(len(entities))

	for i, e := range entities {
		m.AddEntityUpdate(e, i)
	}
}

// Adds a single entity to to the update message. This will
// auto grow the message as needed. The second parameter is
// the the index to insert the info into the message. if it is -1
// the message will be appended to the end of the message. If the 
// idx is provided a check for message size is nto made.
func (m *MsgGameUpdate) AddEntityUpdate(e *Entity, idx int) {
	i := idx
	if i == -1 {
		i = len(m.Es)
		m.growEntityUpdatesToFit(1)
	}

	m.Es[i].Id = uint64(e.id)
	m.Es[i].T = int(e.typ)
	m.Es[i].St = int(e.state)
	m.Es[i].Ttl = int64(e.ttl / time.Millisecond)
	m.Es[i].CAt = e.createdAt.Unix()
	m.Es[i].UAt = e.updatedAt.Unix()
	m.Es[i].X = e.pos.x
	m.Es[i].Y = e.pos.y
	m.Es[i].W = e.pos.width
	m.Es[i].H = e.pos.height
	m.Es[i].C = e.color
}
