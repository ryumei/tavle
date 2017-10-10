package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

/*
func WriteMessage(conn *websocket.Conn, message string) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func ReadMessage(conn *websocket.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	if messageType != websocket.TextMessage {
		return "", errors.New("invalid message type")
	}
	return string(p), nil
}
*/

func TestValidCase(t *testing.T) {
	// Create server
	ts := httptest.NewServer(http.HandlerFunc(handleConnections))
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

	expected := Message{
		Email:    "room1",
		Username: "John",
		Message:  "my message",
		Room:     "foyer",
	}
	payload, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to create json payload. %v", err)
	}

	if err := WriteMessage(client1, string(payload)); err != nil {
		t.Fatalf("Failed to send message. %v", err)
	}

	/*if err := ws.WriteMessage(client2, `join {"name":"room1"}`); err != nil {
		t.Fatalf("Failed to send message. %v", err)
	}*/

	res, err := ReadMessage(client2)
	if err != nil {
		t.Error(err)
	}

	result := new(Message)
	if err = json.Unmarshal([]byte(res), &result); err != nil {
		t.Error(err)
	}

	if expected != *result {
		t.Errorf("Response is not valid '%s' <> '%s'", expected, *result)
	}
}
