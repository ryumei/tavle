package main

import (
	"time"
)

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
	{"2018-01-23T17:34:58+09:00", 1516696498000000, []byte{128, 208, 254, 107, 109, 99, 5, 0}},
}

var genkeyTests = []struct {
	descrption string
	timestamp  time.Time
	username   string
	bytes      []byte
}{
	{
		"Zero",
		time.Unix(0, 0),
		"user",
		[]byte{
			0, 0, 0, 0, 0, 0, 0, 0,
			238, 17, 203, 177, 144, 82, 228, 11,
			 7,170, 192, 202, 6, 12, 35, 238,
		},
	},
	{
		"2018-01-23T17:34:58+09:00",
		time.Unix(0, 1516696498000000),
		"admin",
		[]byte{
			128, 208, 254, 107, 109, 99, 5, 0,
			33, 35, 47, 41, 122, 87, 165, 167,
			67, 137, 74, 14, 74, 128, 31, 195,
		},
	},
}
