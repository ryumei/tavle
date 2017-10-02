package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// https://qiita.com/nirasan/items/330a0e23f3877bce0051

func TestValidCase(t *testing.T) {
	// Create server
	ts := httptest.NewServer(http.HandlerFunc(handleConnections))
	defer ts.Close()

	client1, err := WSClient(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client1.Close()

	client2, err := WSClient(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client2.Close()

	expected := Message{
		Email:    "room1",
		Username: "John",
		Message:  "my message",
		Room:     "my room",
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
