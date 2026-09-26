package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// NotificationConfig is configuration for notifications to other services
type NotificationConfig struct {
	// SlackWebhooks maps a room name to a Slack incoming webhook URL
	SlackWebhooks map[string]string
}

var slackClient = &http.Client{Timeout: 10 * time.Second}

// notifySlack posts that a new message arrived in the room to the webhook of the room
func notifySlack(room string) {
	url, ok := conf.Notification.SlackWebhooks[room]
	if !ok || url == "" {
		return
	}
	payload, err := json.Marshal(map[string]string{
		"text": fmt.Sprintf("New message in tavle room '%s'", room),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to marshaling a slack message %v", err)
		return
	}
	// The webhook URL is secret. Do not log it.
	resp, err := slackClient.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("[WARN] Failed to notify slack for room '%s'", room)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARN] Failed to notify slack for room '%s': %s", room, resp.Status)
	}
}
