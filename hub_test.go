package main

import "testing"

func TestValidateRoomName(t *testing.T) {
	for _, item := range roomnameTests {
		if got := validateRoomname(item.roomname); got != item.want {
			t.Fatalf("validateRoomname(%s) want %s", item.roomname, item.want)
		}
	}
}
