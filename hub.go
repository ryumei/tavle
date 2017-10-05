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

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered connected clients in rooms.
	rooms map[string]map[*Subscription]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan Subscription

	// Unregister requests from clients.
	unregister chan Subscription
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan Subscription),
		unregister: make(chan Subscription),
		rooms:      make(map[string]map[*Subscription]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case sub := <-h.register:
			log.Printf("[DEBUG] hub register %s", sub.room)
			connections := h.rooms[sub.room]
			if connections == nil {
				log.Printf("[DEBUG] Create a new room %s", sub.room)
				connections = make(map[*Subscription]bool)
				h.rooms[sub.room] = connections
			}
			connections[&sub] = true
		case sub := <-h.unregister:
			log.Printf("[DEBUG] hub unregister")

			connections := h.rooms[sub.room]
			if connections != nil {
				if _, ok := connections[&sub]; ok {
					delete(connections, &sub)
					close(sub.send)
					if len(connections) == 0 { // Close a room
						delete(h.rooms, sub.room)
					}
				}
			}
		case msg := <-h.broadcast:
			log.Printf("[DEBUG] hub boradcast to room:%s", msg.Room) // from readPump
			connections := h.rooms[msg.Room]
			log.Printf("[DEBUG] # of connections %d", len(connections)) // from readPump

			for sub := range connections {
				rawMessage, err := json.Marshal(msg)
				if err != nil {
					log.Printf("[ERROR] Failed to marshaling %v", msg)
				}
				select {
				case sub.send <- rawMessage:
					//case sub.send <- []byte(msg.Message):
					log.Printf("[DEBUG] hub send")

					//TODO switch room
					log.Printf("[DEBUG] room: %s", sub.room)
					log.Printf("[DEBUG] message: %s", msg.Message)
					//TODO?
				default:
					log.Printf("[DEBUG] hub default close connection")
					close(sub.send)
					delete(connections, sub)
					if len(connections) == 0 { // Close a room
						delete(h.rooms, msg.Room)
					}
				}
			}
		}
	}
}
