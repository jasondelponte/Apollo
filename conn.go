package main

import (
	gnws "code.google.com/p/go.net/websocket"
	"encoding/json"
	gbws "github.com/garyburd/go-websocket/websocket"
	"io"
	"io/ioutil"
	"log"
	"time"
)

type Connection interface {
	GetId() uint64
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
func NewGbWsConn(id uint64, ws *gbws.Conn) *GbWsConn {
	return &GbWsConn{
		id:   id,
		send: make(chan []byte, 256),
		ws:   ws,
	}
}

type GbWsConn struct {
	id uint64

	// The websocket connection.
	ws *gbws.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Buffered channel for readers
	reader chan MessageIn
}

// Returns the connection's id
func (c GbWsConn) GetId() uint64 {
	return c.id
}

// Sets the channel the connection should forward incomming messages to
func (c *GbWsConn) AttachReader(reader chan MessageIn) {
	c.reader = reader
}

// Serializes an object and sends it across the the wire
func (c *GbWsConn) Send(msg interface{}) error {
	marshaled, err := json.Marshal(msg)
	if err != nil {
		log.Println("ERROR", "Connection", c.id, "Failed to marshal data to send to client")
		return err
	}
	c.send <- marshaled
	return nil
}

// Closes the send and attached reader channels
func (c *GbWsConn) Close() {
	close(c.send)
	close(c.reader)
}

// Handler furnction for periodic reading from the input socks.
// Handles closing of the socket, of the connection drops
func (c *GbWsConn) ReadPump() {
	defer c.ws.Close()
	for {
		c.ws.SetReadDeadline(time.Now().Add(readWait))
		op, r, err := c.ws.NextReader()
		if err != nil {
			log.Println("Error getting next reader", err)
			return
		}
		if op != gbws.OpText {
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
			c.ws.WriteControl(gbws.OpClose,
				gbws.FormatCloseMessage(gbws.CloseMessageTooBig, ""),
				time.Now().Add(time.Second))
			return
		}

		var unmarshaled MessageIn
		json.Unmarshal(message, &unmarshaled)

		c.reader <- unmarshaled
	}
}

// write writes a message with the given opCode and payload.
func (c *GbWsConn) write(opCode int, payload []byte) error {
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
func (c *GbWsConn) WritePump() {
	defer c.ws.Close()
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(gbws.OpClose, []byte{})
				return
			}
			if err := c.write(gbws.OpText, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(gbws.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

func NewGnWsConn(id uint64, ws *gnws.Conn) *GnWsConn {
	return &GnWsConn{
		id:    id,
		send:  make(chan interface{}, 256),
		ws:    ws,
		alive: true,
	}
}

type GnWsConn struct {
	id     uint64
	reader chan MessageIn
	send   chan interface{}
	ws     *gnws.Conn
	alive  bool
}

// Returns the connection's id
func (c GnWsConn) GetId() uint64 {
	return c.id
}

func (c *GnWsConn) AttachReader(reader chan MessageIn) {
	c.reader = reader
}
func (c *GnWsConn) Send(msg interface{}) error {
	c.send <- msg
	return nil
}
func (c *GnWsConn) ReadPump() {
	for {
		var msg MessageIn
		err := gnws.JSON.Receive(c.ws, &msg)
		if err != nil {
			log.Println("Failed to read from ws, ", err, "id", c.id)
			return
		}

		c.reader <- msg
	}
}
func (c *GnWsConn) WritePump() {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				log.Println("WS send chan closed id", c.id)
				return
			}

			err := gnws.JSON.Send(c.ws, msg)
			if err != nil {
				log.Println("Failed to write to ws, ", err, "id", c.id)
				return
			}
		}
	}
}
func (c *GnWsConn) Close() {
	if c.alive {
		close(c.send)
		close(c.reader)
		c.alive = false
	} else {
		log.Println("trying to close already closed connection")
	}
}
