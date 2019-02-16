package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Client struct {
	id     string
	conn   *websocket.Conn
	sendCh chan struct {
		messageType int
		payload     []byte
	}

	// references to room's channels
	broadcastToRoomCh chan struct {
		from    string
		message struct {
			messageType int
			payload     []byte
		}
	}
	unregisterToRoomCh chan struct{}

	isClosed bool
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		id:   GenerateClientID(),
		conn: conn,
		sendCh: make(chan struct {
			messageType int
			payload     []byte
		}, 256),
		isClosed: false,
	}
}

func (c *Client) run() {
	fmt.Printf("Client %s connected\n", c.id)
	go c.readPump()
	go c.writePump()
}

func (c *Client) readPump() {
	for !c.isClosed {
		mt, payload, err := c.conn.ReadMessage()
		if err != nil {
			// this if means that the other end closed, if the other end closes we need to
			// close all client
			// but if the error is not from the other side, that means that `the room` sent
			// a signal to close the client
			if websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseProtocolError,
				websocket.CloseUnsupportedData,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
				websocket.CloseInvalidFramePayloadData,
				websocket.ClosePolicyViolation,
				websocket.CloseMessageTooBig,
				websocket.CloseMandatoryExtension,
				websocket.CloseInternalServerErr,
				websocket.CloseServiceRestart,
				websocket.CloseTryAgainLater,
				websocket.CloseTLSHandshake,
			) {
				if c.unregisterToRoomCh != nil {
					c.unregisterToRoomCh <- struct{}{}
				}
			}

			break
		}

		if c.broadcastToRoomCh != nil {
			c.broadcastToRoomCh <- struct {
				from    string
				message struct {
					messageType int
					payload     []byte
				}
			}{
				from: c.id,
				message: struct {
					messageType int
					payload     []byte
				}{mt, payload},
			}
		}
	}
}

func (c *Client) writePump() {
loop:
	for !c.isClosed {
		select {
		case msg, ok := <-c.sendCh:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				break loop
			}

			err := c.conn.WriteMessage(msg.messageType, msg.payload)
			if err != nil {
				break loop
			}
		}
	}
}

func (c *Client) close() {
	c.conn.Close()
	if !c.isClosed {
		close(c.sendCh)
	}
	c.isClosed = true

	fmt.Printf("Client %s disconnected\n", c.id)
}
