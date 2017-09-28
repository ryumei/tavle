package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryumei/tavle/websocket"
)

// https://qiita.com/nirasan/items/330a0e23f3877bce0051

func TestValidCase(t *testing.T) {
	// Create server
	ts := httptest.NewServer(http.HandlerFunc(handleConnections))
	defer ts.Close()

	client1, err := ws.Client(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client1.Close()

	client2, err := ws.Client(ts)
	if err != nil {
		t.Fatalf("Failed to create a client. %v", err)
	}
	defer client2.Close()

	payload := `{"email":"room1","username":"room1","message":"my message"}`

	if err := ws.WriteMessage(client1, payload); err != nil {
		t.Fatalf("Failed to send message. %v", err)
	}

	/*if err := ws.WriteMessage(client2, `join {"name":"room1"}`); err != nil {
		t.Fatalf("Failed to send message. %v", err)
	}*/

	res, err := ws.ReadMessage(client2)
	if err != nil {
		t.Error(err)
	}
	if res != payload {
		t.Errorf("Response is not valid '%v' != '%v'", payload, res)
	}
}
