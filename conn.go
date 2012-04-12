package main

import (
	"encoding/json"
	"github.com/garyburd/go-websocket/websocket"
	"io"
	"io/ioutil"
	"log"
	"time"
)

type Connection interface {
	AttachReader(reader chan MessageIn)
	Send(msg interface{}) error
	ReadPump()
	WritePump()
	Close()
}

type MessageIn struct {
	ReqId string
	Cmd   *MsgPartCommand
}

type MsgPartCommand struct {
	OpCode string
}

type MsgResponse struct {
	Rsp string
}

type MsgBlock struct {
	X int
	Y int
	R int
	G int
	B int
	A int
	W int
	H int
}

const (
	readWait       = 60 * time.Second
	pingPeriod     = 25 * time.Second
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

// Creates a new instance of the ws
func NewWsConn() *WsConn {
	return &WsConn{}
}

type WsConn struct {
	id uint64

	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Buffered channel for readers
	reader chan MessageIn
}

// Sets the channel the connection should forward incomming messages to
func (c *WsConn) AttachReader(reader chan MessageIn) {
	c.reader = reader
}

// Serializes an object and sends it across the the wire
func (c *WsConn) Send(msg interface{}) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		log.Println("ERROR", "Connection", c.id, "Failed to marshal data to send to client")
		return err
	}
	c.send <- marshaled
	return nil
}

// Closes the send and attached reader channels
func (c *WsConn) Close() {
	close(c.send)
	close(c.reader)
}

// Handler furnction for periodic reading from the input socks.
// Handles closing of the socket, of the connection drops
func (c *WsConn) ReadPump() {
	defer c.ws.Close()
	for {
		c.ws.SetReadDeadline(time.Now().Add(readWait))
		op, r, err := c.ws.NextReader()
		if err != nil {
			log.Println("Error getting next reader", err)
			return
		}
		if op != websocket.OpText {
			continue
		}
		lr := io.LimitedReader{R: r, N: maxMessageSize + 1}
		message, err := ioutil.ReadAll(&lr)
		if err != nil {
			log.Println("Error reading message", err)
			return
		}
		if lr.N <= 0 {
			log.Println("No data received from message, closing")
			c.ws.WriteControl(websocket.OpClose,
				websocket.FormatCloseMessage(websocket.CloseMessageTooBig, ""),
				time.Now().Add(time.Second))
			return
		}

		var unmarshaled MessageIn
		json.Unmarshal(message, &unmarshaled)

		c.reader <- unmarshaled
	}
}

// write writes a message with the given opCode and payload.
func (c *WsConn) write(opCode int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	w, err := c.ws.NextWriter(opCode)
	if err != nil {
		log.Println("Failed to get next writer", err)
		return err
	}
	if _, err := w.Write(payload); err != nil {
		log.Println("Failed to write payload", err)
		w.Close()
		return err
	}
	return w.Close()
}

// writePump pumps messages from the hub to the websocket connection.
func (c *WsConn) WritePump() {
	defer c.ws.Close()
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.OpClose, []byte{})
				return
			}
			if err := c.write(websocket.OpText, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}
