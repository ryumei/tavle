package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"math/rand"
	"testing"
	"time"
)

func TestTime2Bytes(t *testing.T) {
	for _, item := range time2BytesTests {
		expected := item.bytes
		data := time.Unix(0, item.timestampNs)
		result := time2bytes(data)

		if !bytes.Equal(result, expected) {
			t.Fatalf("[ERROR] %v <> %v for %v", result, item.bytes, item.timestampNs)
		}
	}
}

func TestBytes2Time(t *testing.T) {
	for _, item := range time2BytesTests {
		expected := item.timestampNs
		data := item.bytes
		result := bytes2time(data).UnixNano()

		if result != expected {
			t.Fatalf("[ERROR] %v <> %v for %v", result, item.bytes, item.timestampNs)
		}
	}
}

func TestDBKey(t *testing.T) {
	for _, item := range genkeyTests {
		expected := item.bytes
		result := dbKeySHA1(item.timestamp, item.username)
		if len(result) != 28 {
			t.Fatalf("Invalid hash length %d", len(result))
		}

		if !bytes.Equal(result, expected) {
			t.Fatalf("[ERROR] result %v <> expected %v", result, expected)
		}

		sameResult := dbKeySHA1(item.timestamp, item.username)
		if !bytes.Equal(result, sameResult) {
			t.Fatalf("[ERROR] result1 %v <> result2 %v", result, sameResult)
		}

		anotherResult := dbKeySHA1(item.timestamp, item.username+"another")
		if bytes.Equal(result, anotherResult) {
			t.Fatalf("[ERROR] result1 %v == result2 %v", result, anotherResult)
		}
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

/*
 *
 * BenchmarkDBKeyMD5-8      	 5000000	       457 ns/op
 * BenchmarkDBKeySHA1-8     	 3000000	       539 ns/op
 * BenchmarkDBKeySHA256-8   	 2000000	       772 ns/op
 */

// just for reference
func dbKeyMD5(t time.Time, user string) []byte {
	var k = make([]byte, 0, 8+md5.Size)
	for _, c := range time2bytes(t) {
		k = append(k, c)
	}
	for _, c := range md5.Sum([]byte(user)) {
		k = append(k, c)
	}
	return k
}

// just for reference
func dbKeySHA256(t time.Time, user string) []byte {
	var k = make([]byte, 0, 8+sha256.Size)
	for _, c := range time2bytes(t) {
		k = append(k, c)
	}
	for _, c := range sha256.Sum256([]byte(user)) {
		k = append(k, c)
	}
	return k
}

func BenchmarkDBKeyMD5(b *testing.B) {
	username := RandStringBytes(32)
	for i := 0; i < b.N; i++ {
		dbKeyMD5(time.Now(), username)
	}
}

func BenchmarkDBKeySHA1(b *testing.B) {
	username := RandStringBytes(32)
	for i := 0; i < b.N; i++ {
		dbKeySHA1(time.Now(), username)
	}
}

func BenchmarkDBKeySHA256(b *testing.B) {
	username := RandStringBytes(32)
	for i := 0; i < b.N; i++ {
		dbKeySHA256(time.Now(), username)
	}
}

func TestGetDatabase(t *testing.T) {
	dataDir := "test"
	roomname := "testroom"

	db, err := GetWritableDB(dataDir, roomname)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
}

func TestReadPost(t *testing.T) {
	var roomname = "testroom"
	var dataDir = "test"
	var secret = []byte("CHANGEME_16CHARS") // AWS restriction

	msg := Message{
		Email:     "test@tavle.example.com",
		Username:  "Tavle Test",
		Message:   "テストメッセージ",
		Room:      roomname,
		Timestamp: time.Now(),
	}

	SavePost(msg, dataDir, secret)

	posts, err := LoadPosts(
		roomname,
		time.Now(),
		86400,
		dataDir,
		secret,
	)
	if err != nil {
		t.Fatalf("Load post error %v", err)
	}

	if len(posts) < 1 {
		t.Fatalf("No posts found.")
	}

}
