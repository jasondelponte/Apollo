package main

const (
	UPDATE_TYPE_ADD   = 1
	ENTITY_TYPE_BLOCK = 0
)

// Incoming message container which contains all possible 
// input message types
type MessageIn struct {
	ReqId string
	Cmd   *MsgPartCommand
}

type MsgPartCommand struct {
	OpCode int
}

// Response messages
type MsgBoardCurrent struct {
	E []interface{} // Entities
}
type MsgBoardUpdates struct {
	BU []*MsgBoardUpdateItem // Board Updates
}
type MsgBoardUpdateItem struct {
	T int         // Update Type
	E interface{} // Entity
}
type MsgBlock struct {
	T int
	X int
	Y int
	R int
	G int
	B int
	A int
	W int
	H int
}
