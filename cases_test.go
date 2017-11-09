package main

var roomnameTests = []struct {
	description string
	roomname    string
	want        string
}{
	{"default roomname", DefaultRoomname, DefaultRoomname},
	{"blank roomname", "", DefaultRoomname},
	{"with blank", "test room", DefaultRoomname},
	{"with hyphen", "test-room", "test-room"},
	{"with underscore", "test_room", "test_room"},
	{"with dot", "test.room", "test.room"},
	{"with number ", "testroom1", "testroom1"},
}
