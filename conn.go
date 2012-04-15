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

type ConnError struct {
	ConnErrorString string
}

func (c *ConnError) Error() string { return c.ConnErrorString }

var (
	ConnErrorSendClosed = &ConnError{"Connection's send chan already closed"}
	ConnErrorReadClosed = &ConnError{"Connection's reader chan already closed"}
)

type Connection interface {
	GetId() uint64
	AttachReader(reader chan MessageIn)
	Send(msg interface{}) error
	ReadPump()
	WritePump()
	Close()
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
	if c.send == nil {
		return ConnErrorSendClosed
	}

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
	if c.send != nil {
		close(c.send)
	}
	if c.reader != nil {
		close(c.reader)
	}

	c.send = nil
	c.reader = nil
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

		if c.reader == nil {
			log.Println("GbWsConn failed to send to reader because chan closed")
			return
		}

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

// Create a new instance of the go.net websocket
func NewGnWsConn(id uint64, ws *gnws.Conn) *GnWsConn {
	return &GnWsConn{
		id:   id,
		send: make(chan interface{}, 256),
		ws:   ws,
	}
}

// Connection object for use with the go.net websocket
type GnWsConn struct {
	id     uint64
	reader chan MessageIn
	send   chan interface{}
	ws     *gnws.Conn
}

// Returns the connection's id
func (c GnWsConn) GetId() uint64 {
	return c.id
}

// Adds a new reader to the go.net websocket connection
// This channel will be notified when a message is received
// from a connection.
func (c *GnWsConn) AttachReader(reader chan MessageIn) {
	c.reader = reader
}

// Sends an object which will be serialized and sent to the connection
func (c *GnWsConn) Send(msg interface{}) error {
	if c.send == nil {
		return ConnErrorSendClosed
	}

	c.send <- msg
	return nil
}

// Read event loop, terminates when read from client fails
func (c *GnWsConn) ReadPump() {
	defer func() { log.Println("Connection ", c.id, "read pump terminating") }()
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

// Write event loop, termintes when writes to the client fails, or the send to conn
// channel is closed.
func (c *GnWsConn) WritePump() {
	defer func() { log.Println("Connection ", c.id, "write pump terminating") }()
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

// Closes all channels on the connection, and invalidates
// their reference.
func (c *GnWsConn) Close() {
	if c.send != nil {
		close(c.send)
	}
	if c.reader != nil {
		close(c.reader)
	}

	c.send = nil
	c.reader = nil
}
