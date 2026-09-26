package main

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestSampleConfig(t *testing.T) {
	var c config
	if _, err := toml.DecodeFile("tavle.tml.sample", &c); err != nil {
		t.Fatal(err)
	}
	if len(c.Notification.SlackWebhooks) != 0 {
		t.Fatalf("[ERROR] no webhooks expected in the sample %v", c.Notification.SlackWebhooks)
	}
}

func TestSlackWebhooksConfig(t *testing.T) {
	var c config
	text := `
[Notification.SlackWebhooks]
foyer = "https://example.com/foyer"
"test.room" = "https://example.com/test.room"
`
	if _, err := toml.Decode(text, &c); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"foyer":     "https://example.com/foyer",
		"test.room": "https://example.com/test.room",
	}
	for room, url := range want {
		if got := c.Notification.SlackWebhooks[room]; got != url {
			t.Fatalf("[ERROR] room %s want %s got %s", room, url, got)
		}
	}
}
