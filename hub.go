package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

	// Type is "delete" for a request to delete a message
	Type string `json:"type,omitempty"`
	// ID identifies a message
	ID string `json:"id,omitempty"`
	// DeleteToken is required to delete a message. It is given only to the sender.
	DeleteToken string `json:"deleteToken,omitempty"`

	// sender is the connection which sent the message
	sender *connection
}

// deleteMessage notifies a message is deleted
type deleteMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// deleteToken returns the token to delete the message in the room
func deleteToken(room string, id string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(room + "/" + id))
	return hex.EncodeToString(mac.Sum(nil))
}

// validDeleteToken reports whether the token is valid to delete the message
func validDeleteToken(room string, id string, token string, secret []byte) bool {
	return hmac.Equal([]byte(token), []byte(deleteToken(room, id, secret)))
}

// membersMessage notifies the number of connections in a room
type membersMessage struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
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
			h.notifyMembers(roomname)
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
					h.notifyMembers(sub.room)
				}
			}
		case msg := <-h.broadcast:
			if msg.Type == "delete" {
				h.deletePost(msg)
				continue
			}
			msg.Timestamp = time.Now()
			msg.ID = messageID(msg.Timestamp, msg.Username)
			log.Printf("[DEBUG] hub boradcast to room:%s", msg.Room) // called from readPump
			connections := h.rooms[msg.Room]
			log.Printf("[DEBUG] # of connections %d", len(connections))

			rawMessage, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[ERROR] Failed to marshaling a message '%v'", msg)
			}
			// Only the sender gets the token to delete the message
			own := msg
			own.DeleteToken = deleteToken(msg.Room, msg.ID, dbSecret)
			rawOwnMessage, err := json.Marshal(own)
			if err != nil {
				log.Printf("[ERROR] Failed to marshaling a message '%v'", own)
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

			removed := false
			for c := range connections {
				raw := rawMessage
				if c == msg.sender {
					raw = rawOwnMessage
				}
				select {
				case c.send <- raw:
					log.Printf("[DEBUG] hub send [%s]: %s", msg.Room, msg.Message)
					if len(rawAdminMessage) > 0 {
						c.send <- rawAdminMessage
					}
				default:
					log.Printf("[DEBUG] hub default close connection")
					close(c.send)
					delete(connections, c)
					removed = true
					if len(connections) == 0 { // Close a room
						delete(h.rooms, msg.Room)
					}
				}
			}
			if removed {
				h.notifyMembers(msg.Room)
			}
		}
	}
}

// notifyMembers sends the number of connections to all connections in the room
func (h *Hub) notifyMembers(roomname string) {
	rawMessage, err := json.Marshal(membersMessage{Type: "members", Count: len(h.rooms[roomname])})
	if err != nil {
		log.Printf("[ERROR] Failed to marshaling a members message %v", err)
		return
	}
	h.notify(roomname, rawMessage)
}

// deletePost deletes the message and notifies it to all connections in the room
func (h *Hub) deletePost(msg Message) {
	rawMessage, err := json.Marshal(deleteMessage{Type: "delete", ID: msg.ID})
	if err != nil {
		log.Printf("[ERROR] Failed to marshaling a delete message %v", err)
		return
	}
	writer <- msg
	h.notify(msg.Room, rawMessage)
}

// notify sends the message to all connections in the room without blocking
func (h *Hub) notify(roomname string, rawMessage []byte) {
	for c := range h.rooms[roomname] {
		select {
		case c.send <- rawMessage:
		default:
			log.Printf("[DEBUG] hub skip a notification to a busy connection")
		}
	}
}

var roomnameMatch = regexp.MustCompile(`^[\w\-\.]+$`)

func sanitizeRoomname(roomname string) string {
	if roomnameMatch.Match([]byte(roomname)) {
		return roomname
	}
	log.Printf("[WARN] Use default roomname '%s' instead of '%s'.", DefaultRoomname, roomname)
	return DefaultRoomname
}
