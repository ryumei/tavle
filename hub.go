// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Message is message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
	// TODO room id
}

type MessageEnvelope struct {
	data []byte
	room string
}

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered connected clients in rooms.
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan MessageEnvelope

	// Register requests from the clients.
	register chan Subscription

	// Unregister requests from clients.
	unregister chan Subscription
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan MessageEnvelope),
		register:   make(chan Subscription),
		unregister: make(chan Subscription),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case s := <-h.register:
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*Client]bool)
				h.rooms[s.room] = connections
			}
			connections[s.conn] = true
		case s := <-h.unregister:
			connections := h.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 { // Close a room
						delete(h.rooms, s.room)
					}
				}
			}
		case message := <-h.broadcast:
			connections := h.rooms[message.room]
			for client := range connections {
				select {
				case client.send <- message.data: //TODO messageenvelop (with roomname)
				default:
					close(client.send)
					delete(connections, client)
					if len(connections) == 0 {
						delete(h.rooms, message.room)
					}
				}
			}
		}
	}
}
