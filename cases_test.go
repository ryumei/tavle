package main

var roomnameTests = []struct {
	description string
	roomname    string
	want        bool
}{
	{"default roomname", DefaultRoomname, true},
	{"blank roomname", "", false},
	{"with blank", "test room", false},
	{"with hyphen", "test-room", true},
	{"with underscore", "test_room", true},
	{"with dot", "test.room", true},
	{"with number ", "testroom1", true},
}
