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

var time2BytesTests = []struct {
	descrption  string
	timestampNs int64
	bytes       []byte
}{
	{"Zero", 0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
	{"arbitrary time", 1516696498000000, []byte{128, 208, 254, 107, 109, 99, 5, 0}},
}
