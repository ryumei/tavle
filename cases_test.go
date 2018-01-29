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
	{
		"2018-01-23T17:34:58+09:00",
		1516696498000000,
		[]byte{0, 5, 99, 109, 107, 254, 208, 128},
	},
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
			18, 222, 169, 111, 236, 32, 89, 53,
			102, 171, 117, 105, 44, 153, 73, 89,
			104, 51, 173, 201,
		},
	},
	{
		"2018-01-23T17:34:58+09:00",
		time.Unix(0, 1516696498000000),
		"admin",
		[]byte{
			0, 5, 99, 109, 107, 254, 208, 128,
			208, 51, 226, 42, 227, 72, 174, 181,
			102, 15, 194, 20, 10, 236, 53, 133,
			12, 77, 169, 151,
		},
	},
}

// crypto_test
var encryptionTests = []struct {
	description string
	data        string
	secret      string
}{
	{"Zero", "", "CHANGEME_16CHARS"},
	{"arbitrary", "My message", "CHANGEME_16CHARS"},
}
