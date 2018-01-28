package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// https://qiita.com/nirasan/items/330a0e23f3877bce0051

// wsClient is test websocket client
func wsClient(ts *httptest.Server) (*websocket.Conn, error) {
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
		log.Printf("[WARN] %v", err)
		return "", err
	}
	if messageType != websocket.TextMessage {
		log.Printf("[WARN] Invalid message type %v", messageType)
		return "", errors.New("invalid message type")
	}
	return string(p), nil
}

func TestValidCase(t *testing.T) {
	go hub.run()

	// Create server
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			serveWs(&hub, w, r)
		}))
	defer ts.Close()

	client1, err := wsClient(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client1.Close()

	client2, err := wsClient(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client2.Close()

	// 0. Welcome message

	/*

		expected0 := Message{
			Email:     "admin@tavle.example.com",
			Username:  "Tavle Admin",
			Message:   "Welcome to room 'foyer'",
			Room:      "foyer",
			Timestamp: time.Now(),
		}
	*/
	/*
		res, err := ReadMessage(client2)
		if err != nil {
			t.Error(err)
		}
		result := new(Message)
		if err = json.Unmarshal([]byte(res), &result); err != nil {
			t.Error(err)
		}
		if expected0.Email != result.Email {
			t.Errorf("Response is not valid '%s' <> '%s'", expected0.Email, result.Email)
		}
		if expected0.Username != result.Username {
			t.Errorf("Response is not valid '%s' <> '%s'", expected0.Username, result.Username)
		}
		if expected0.Room != result.Room {
			t.Errorf("Response is not valid '%s' <> '%s'", expected0.Room, result.Room)
		}
		if expected0.Message != result.Message {
			t.Errorf("Response is not valid '%s' <> '%s'", expected0.Message, result.Message)
		}
	*/
	// Send and receive

	expected := Message{
		Email:     "room1",
		Username:  "John",
		Message:   "my message",
		Room:      "foyer",
		Timestamp: time.Now(),
	}
	payload, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to create json payload. %v", err)
	}

	// Flush connection
	for {
		_, err := ReadMessage(client2)
		if err != nil {
			log.Print(err)
			break
		}
	}

	if err := WriteMessage(client2, string(payload)); err != nil {
		t.Fatalf("Failed to send message. %v", err)
	}
	time.Sleep(1000)

	res, err := ReadMessage(client2)
	if err != nil {
		t.Error(err)
	}
	result := new(Message)
	if err = json.Unmarshal([]byte(res), &result); err != nil {
		log.Println(res) //###############
		t.Error(err)
	}
	if expected != *result {
		t.Errorf("Response is not valid '%s' <> '%s'", expected, *result)
	}

}
