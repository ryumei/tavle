package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"
)

// Message is a message object
type Message struct {
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Room      string    `json:"room"`
	Timestamp time.Time `json:"timestamp"`
}

// subscription is connection and joined room
type subscription struct {
	conn *connection
	room string
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered connected clients in rooms.
	rooms map[string]map[*connection]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan subscription

	// Unregister requests from clients.
	unregister chan subscription
}

// DefaultRoomname 省略時のルーム名
var DefaultRoomname = "foyer"

// hub is the global hub
var hub = Hub{
	broadcast:  make(chan Message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func (h *Hub) run() {
	log.Printf("[DEBUG] hub run enter")
	for {
		select {
		case sub := <-h.register:
			roomname := sub.room
			log.Printf("[DEBUG] hub register room '%s'", sub.room)
			connections := h.rooms[roomname]
			if connections == nil {
				log.Printf("[DEBUG] Create a new room '%s'", roomname)
				connections = make(map[*connection]bool)
				h.rooms[roomname] = connections
			}
			connections[sub.conn] = true
		case sub := <-h.unregister:
			log.Printf("[DEBUG] hub unregister")
			connections := h.rooms[sub.room]
			if connections != nil {
				if _, ok := connections[sub.conn]; ok {
					delete(connections, sub.conn)
					close(sub.conn.send)
					if len(connections) == 0 { // Close a room
						delete(h.rooms, sub.room)
					}
				}
			}
		case msg := <-h.broadcast:
			msg.Timestamp = time.Now()
			log.Printf("[DEBUG] hub boradcast to room:%s", msg.Room) // called from readPump
			connections := h.rooms[msg.Room]
			log.Printf("[DEBUG] # of connections %d", len(connections))

			rawMessage, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[ERROR] Failed to marshaling a message '%v'", msg)
			}
			writer <- msg

			var rawAdminMessage []byte
			if strings.HasPrefix(msg.Message, "admin ") {
				admMsg := Message{
					Email:     "",
					Username:  "Tavle Admin",
					Message:   GetQuote(),
					Room:      DefaultRoomname,
					Timestamp: time.Now(),
				}
				rawAdminMessage, _ = json.Marshal(admMsg)
				writer <- admMsg
			}

			for c := range connections {
				select {
				case c.send <- rawMessage:
					log.Printf("[DEBUG] hub send [%s]: %s", msg.Room, msg.Message)
					if len(rawAdminMessage) > 0 {
						c.send <- rawAdminMessage
					}
				default:
					log.Printf("[DEBUG] hub default close connection")
					close(c.send)
					delete(connections, c)
					if len(connections) == 0 { // Close a room
						delete(h.rooms, msg.Room)
					}
				}
			}
		}
	}
}

var roomnameMatch = regexp.MustCompile(`^[\w\-\.]+$`)

func sanitizeRoomname(roomname string) string {
	if roomnameMatch.Match([]byte(roomname)) {
		log.Printf("[WARN] Use default roomname '%s' instead of '%s'.", DefaultRoomname, roomname)
		return roomname
	}
	return DefaultRoomname
}
