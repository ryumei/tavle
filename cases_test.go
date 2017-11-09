package main

var roomnameTests = []struct {
	roomname string
	want     bool
}{
	{"foyer", true},      // default
	{"", false},          //
	{"test room", false}, //
	{"test-room", true},  //
	{"test_room", true},  //
	{"test.room", true},  //
}
