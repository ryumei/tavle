package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// startHubOnce starts the global hub only once since readPump uses it
var startHubOnce sync.Once

// startTestServer starts a websocket server with the global hub
func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	conf.Server.DataDir = t.TempDir()
	dbSecret = []byte("CHANGEME_16CHARS")
	startHubOnce.Do(func() {
		go hub.run()
		go func() {
			for range writer { // drain messages to be saved
			}
		}()
	})

	r := mux.NewRouter()
	r.HandleFunc("/ws/{room}", func(w http.ResponseWriter, r *http.Request) {
		serveWs(&hub, w, r)
	})
	server := httptest.NewServer(r)
	t.Cleanup(server.Close)
	return server
}

func dialRoom(t *testing.T, server *httptest.Server, room string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/" + room
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("[ERROR] %v", err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws
}

// receive returns the first message matched with the condition.
// The connection can not be read any more after it is timed out.
func receive(t *testing.T, ws *websocket.Conn, match func(Message) bool) (Message, bool) {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			return Message{}, false
		}
		var m Message
		json.Unmarshal(raw, &m)
		if match(m) {
			return m, true
		}
	}
}

func TestMessageRoomIsConnectionRoom(t *testing.T) {
	server := startTestServer(t)

	// "bad room" is replaced with DefaultRoomname on the server side
	ws := dialRoom(t, server, "bad%20room")

	post, _ := json.Marshal(Message{Username: "user", Room: "bad room", Message: "hello"})
	if err := ws.WriteMessage(websocket.TextMessage, post); err != nil {
		t.Fatalf("[ERROR] %v", err)
	}

	m, ok := receive(t, ws, func(m Message) bool { return m.Message == "hello" })
	if !ok {
		t.Fatal("[ERROR] own message not received")
	}
	if m.Room != DefaultRoomname {
		t.Fatalf("[ERROR] want room %s got %s", DefaultRoomname, m.Room)
	}
}

func TestDeleteMessage(t *testing.T) {
	server := startTestServer(t)
	sender := dialRoom(t, server, "deleteroom")
	other := dialRoom(t, server, "deleteroom")

	post, _ := json.Marshal(Message{Username: "user", Message: "to be deleted"})
	sender.WriteMessage(websocket.TextMessage, post)
	isPost := func(m Message) bool { return m.Message == "to be deleted" }
	own, ok := receive(t, sender, isPost)
	if !ok || own.ID == "" || own.DeleteToken == "" {
		t.Fatalf("[ERROR] sender should get id and token: %v", own)
	}
	if m, ok := receive(t, other, isPost); !ok || m.ID != own.ID || m.DeleteToken != "" {
		t.Fatalf("[ERROR] others should get id without token: %v", m)
	}

	// A marker post arrives first if the requests with invalid tokens are ignored
	isDelete := func(m Message) bool { return m.Type == "delete" }
	for _, token := range []string{"", "invalid"} {
		req, _ := json.Marshal(Message{Type: "delete", ID: own.ID, DeleteToken: token})
		other.WriteMessage(websocket.TextMessage, req)
	}
	marker, _ := json.Marshal(Message{Username: "other", Message: "marker"})
	other.WriteMessage(websocket.TextMessage, marker)
	if m, ok := receive(t, sender, func(m Message) bool { return isDelete(m) || m.Message == "marker" }); !ok || isDelete(m) {
		t.Fatalf("[ERROR] deleted with invalid token: %v", m)
	}

	req, _ := json.Marshal(Message{Type: "delete", ID: own.ID, DeleteToken: own.DeleteToken})
	sender.WriteMessage(websocket.TextMessage, req)
	for _, ws := range []*websocket.Conn{sender, other} {
		if m, ok := receive(t, ws, isDelete); !ok || m.ID != own.ID {
			t.Fatalf("[ERROR] delete notification not received: %v", m)
		}
	}
}
