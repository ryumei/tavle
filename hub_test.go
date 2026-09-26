package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidateRoomName(t *testing.T) {
	for _, item := range roomnameTests {
		if got := sanitizeRoomname(item.roomname); got != item.want {
			t.Fatalf("[ERROR] %s validateRoomname(\"%s\") want %s",
				item.description, item.roomname, item.want)
		}
	}
}

func newTestHub() *Hub {
	h := &Hub{
		broadcast:  make(chan Message),
		register:   make(chan subscription),
		unregister: make(chan subscription),
		rooms:      make(map[string]map[*connection]bool),
	}
	go h.run()
	return h
}

func receiveMembers(t *testing.T, c *connection) int {
	t.Helper()
	select {
	case raw := <-c.send:
		var m membersMessage
		if err := json.Unmarshal(raw, &m); err != nil || m.Type != "members" {
			t.Fatalf("[ERROR] unexpected message %s", raw)
		}
		return m.Count
	case <-time.After(time.Second):
		t.Fatal("[ERROR] members message not received")
	}
	return -1
}

func TestMembersNotification(t *testing.T) {
	h := newTestHub()
	a := subscription{conn: &connection{send: make(chan []byte, 8)}, room: "room1"}
	b := subscription{conn: &connection{send: make(chan []byte, 8)}, room: "room1"}
	other := subscription{conn: &connection{send: make(chan []byte, 8)}, room: "room2"}

	h.register <- a
	if got := receiveMembers(t, a.conn); got != 1 {
		t.Fatalf("[ERROR] want 1 got %d", got)
	}
	h.register <- other
	if got := receiveMembers(t, other.conn); got != 1 {
		t.Fatalf("[ERROR] other room want 1 got %d", got)
	}
	h.register <- b
	for _, c := range []*connection{a.conn, b.conn} {
		if got := receiveMembers(t, c); got != 2 {
			t.Fatalf("[ERROR] want 2 got %d", got)
		}
	}
	h.unregister <- b
	if got := receiveMembers(t, a.conn); got != 1 {
		t.Fatalf("[ERROR] after unregister want 1 got %d", got)
	}
	select {
	case raw := <-other.conn.send:
		t.Fatalf("[ERROR] other room should not be notified: %s", raw)
	default:
	}
}
