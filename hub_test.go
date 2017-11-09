package main

import "testing"

func TestValidateRoomName(t *testing.T) {
	for _, item := range roomnameTests {
		if got := sanitizeRoomname(item.roomname); got != item.want {
			t.Fatalf("[ERROR] %s validateRoomname(\"%s\") want %s",
				item.description, item.roomname, item.want)
		}
	}
}
