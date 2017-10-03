package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message grom the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Subscription is a client middleman between the websocket connection and the hub.
type Subscription struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages as raw json
	send chan []byte

	// Joined room
	room string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Subscription) readPump() {
	defer func() {
		c.hub.unregister <- *c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	log.Printf("[DEBUG] readPump initiated")
	for {
		_, rawMessage, err := c.conn.ReadMessage()
		log.Printf("[DEBUG] readPump loop, %s", rawMessage) // "send" called
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("[ERROR] %v", err)
			}
			break
		}
		rawMessage = bytes.TrimSpace(bytes.Replace(rawMessage, newline, space, -1))
		var m Message
		json.Unmarshal(rawMessage, &m)
		log.Printf("[DEBUG] %v", m)

		c.hub.broadcast <- m

		/*
			Message{
				Email:    m.Email,
				Username: m.Username,
				Message:  m.Message,
				Room:     m.Room,
			}
		*/
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Subscription) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send: // from client
			log.Printf("[DEBUG] writePump called send '%s'", message)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("[DEBUG] writePump 2")
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("[ERROR] %v", err)
				return
			}
			log.Printf("[DEBUG] writePump 3")
			w.Write(message)
			log.Printf("[DEBUG] writePump 4")

			// c.send???

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				log.Printf("[DEBUG] writePump i%d", i)
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Printf("[ERROR] %v", err)
				return
			}
			log.Printf("[DEBUG] writePump 5")

		case <-ticker.C:
			log.Println("[DEBUG] writePump called ticker")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[DEBUG] serveWs connection upgraded")
	LogRequest(r) //DEBUG

	//TODO get room name
	client := &Subscription{hub: hub, conn: conn, send: make(chan []byte, 256), room: "main"}
	client.hub.register <- *client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump()

}

// WSClient is test websocket client
func WSClient(ts *httptest.Server) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		Subprotocols:    []string{},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	url := strings.Replace(ts.URL, "http://", "ws://", 1)
	header := http.Header{"Accept-Encoding": []string{"gzip"}}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func WriteMessage(conn *websocket.Conn, message string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func ReadMessage(conn *websocket.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	if messageType != websocket.TextMessage {
		return "", errors.New("invalid message type")
	}
	return string(p), nil
}
