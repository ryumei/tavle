package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
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

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Sanitizer sanitize inputs from client browser
var Sanitizer = bluemonday.UGCPolicy()

func sanitize(s string) string {
	return Sanitizer.Sanitize(strings.TrimSpace(s))
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (sub subscription) readPump() {
	conn := sub.conn
	defer func() {
		hub.unregister <- sub
		conn.ws.Close()
	}()
	conn.ws.SetReadLimit(maxMessageSize)
	conn.ws.SetReadDeadline(time.Now().Add(pongWait))
	conn.ws.SetPongHandler(func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	log.Printf("[DEBUG] readPump initiated")
	for {
		_, rawMessage, err := conn.ws.ReadMessage()
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
		m.Message = sanitize(m.Message)

		log.Printf("[DEBUG] unmarshaled message struct %v", m)
		hub.broadcast <- m
	}
}

func (c *connection) write(mt int, payload []byte) error {
	log.Printf("[DEBUG] write a payload")
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (sub *subscription) writePump() {
	conn := sub.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.ws.Close()
	}()
	for {
		select {
		case message, ok := <-conn.send: // from client
			log.Printf("[DEBUG] writePump called send '%s'", message)
			conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, message); err != nil {
				log.Printf("[ERROR] %v", err)
				return
			}
		case <-ticker.C:
			log.Println("[DEBUG] writePump called ticker")
			sub.conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	// var 'room' is expect in the URL path of WebSocket '/ws/{room}'
	vars := mux.Vars(r)
	room := sanitizeRoomname(vars["room"])
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}
	log.Printf("[DEBUG] The connection is upgraded")

	sub := subscription{
		conn: &connection{ws: conn, send: make(chan []byte, 256)},
		room: room,
	}
	hub.register <- sub

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go sub.writePump()
	go sub.readPump()

	welcomMessage := Message{
		Email:    "admin@tavle.example.com",
		Username: "Tavle Admin",
		Message:  fmt.Sprintf("Welcome to room '%s'", room),
		Room:     room,
	}
	rawMessage, err := json.Marshal(welcomMessage)
	if err != nil {
		log.Printf("[WARN] failed unmarshaling %v", err)
	}
	sub.conn.send <- rawMessage
}
