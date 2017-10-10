package main

import "log"
import "encoding/json"

// Message is message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Room     string `json:"room"`
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

func newHub() Hub {
	return Hub{
		broadcast:  make(chan Message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		rooms:      make(map[string]map[*connection]bool),
	}
}

// hub is the global hub
var hub = newHub()

func (h *Hub) run() {
	log.Printf("[DEBUG] hub run enter")
	for {
		select {
		case sub := <-h.register:
			log.Printf("[DEBUG] hub register '%s'", sub.room)
			connections := h.rooms[sub.room]
			if connections == nil {
				log.Printf("[DEBUG] Create a new room %s", sub.room)
				connections = make(map[*connection]bool)
				h.rooms[sub.room] = connections
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
			log.Printf("[DEBUG] hub boradcast to room:%s", msg.Room) // from readPump
			connections := h.rooms[msg.Room]
			log.Printf("[DEBUG] # of connections %d", len(connections))

			for c := range connections {
				rawMessage, err := json.Marshal(msg)
				if err != nil {
					log.Printf("[ERROR] Failed to marshaling %v", msg)
				}
				select {
				case c.send <- rawMessage:
					log.Printf("[DEBUG] hub send [%s]: %s", msg.Room, msg.Message)
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
