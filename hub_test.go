package main

import "testing"

func TestValidateRoomName(t *testing.T) {
	for _, item := range roomnameTests {
		if got := validateRoomname(item.roomname); got != item.want {
			t.Fatalf("[ERROR] validateRoomname(\"%s\") want %t", item.roomname, item.want)
		}
	}
}
