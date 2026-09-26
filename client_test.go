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

func TestMessageRoomIsConnectionRoom(t *testing.T) {
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
	defer server.Close()

	// "bad room" is replaced with DefaultRoomname on the server side
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/bad%20room"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("[ERROR] %v", err)
	}
	defer ws.Close()

	post, _ := json.Marshal(Message{Username: "user", Room: "bad room", Message: "hello"})
	if err := ws.WriteMessage(websocket.TextMessage, post); err != nil {
		t.Fatalf("[ERROR] %v", err)
	}

	ws.SetReadDeadline(time.Now().Add(time.Second))
	for {
		_, raw, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("[ERROR] own message not received: %v", err)
		}
		var m Message
		json.Unmarshal(raw, &m)
		if m.Message != "hello" {
			continue // members notification or history
		}
		if m.Room != DefaultRoomname {
			t.Fatalf("[ERROR] want room %s got %s", DefaultRoomname, m.Room)
		}
		return
	}
}
