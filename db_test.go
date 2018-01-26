package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

/*
 * DB file: room
 *   - date base bucket
 *     - key: timestamp
 * housekeeping -> keep 2 month events in database
 *
 * DB file: single
 *   - Room base bucket
 *     - key: timestamp (RFC3339)
 *
 * DB file: date base
 *   - room base bucket
 *     - key: timestamp
 */

func GetLatestDatabase(dataDir string, roomname string) (*bolt.DB, error) {
	dbFile := fmt.Sprintf("%s-%s.db", roomname, time.Now().Format("2006.01.02"))

	_, err := os.Stat(dataDir)
	if err != nil {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("Failed to created data dir %s. %v", dataDir, err)
		}
	}
	dbPath := filepath.Join(dataDir, dbFile)
	db, err := bolt.Open(dbPath, 0600, nil)
	return db, err
}

/*
 * 64bit の timestamp と
 * ユーザ ID をハッシュした固定長 bytes
 * 順序は ns のタイムスタンプで許してもらおう。
 */

func time2bytes(t time.Time) []byte {
	b := make([]byte, 8)
	ns := t.UTC().UnixNano()
	binary.LittleEndian.PutUint64(b, uint64(ns))
	return b
}

func bytes2time(b []byte) time.Time {
	ns := int64(binary.LittleEndian.Uint64(b))
	return time.Unix(0, ns)
}

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

func key(t time.Time, user string) []byte {
	var k = make([]byte, 0, 8+md5.Size)
	for _, c := range time2bytes(t) {
		k = append(k, c)
	}
	for _, c := range md5.Sum([]byte(user)) {
		k = append(k, c)
	}
	return k
}

func TestKey(t *testing.T) {
	for _, item := range genkeyTests {
		expected := item.bytes
		result := key(item.timestamp, item.username)
		if len(result) != 24 {
			t.Fatalf("Invalid hash length %d", len(result))
		}

		if !bytes.Equal(result, expected) {
			t.Fatalf("[ERROR] result %v <> expected %v", result, expected)
		}

		sameResult := key(item.timestamp, item.username)
		if !bytes.Equal(result, sameResult) {
			t.Fatalf("[ERROR] result1 %v <> result2 %v", result, sameResult)
		}

		anotherResult := key(item.timestamp, item.username+"another")
		if bytes.Equal(result, anotherResult) {
			t.Fatalf("[ERROR] result1 %v == result2 %v", result, anotherResult)
		}

	}
}

func WritePost(m Message) {
	// now under implementing
	dataDir := "."
	db, err := GetLatestDatabase(dataDir, m.Room)
	if err != nil {
		log.Fatal(err)
	}
	bucketName := m.Timestamp.Format("2006.01.02")
	key := time2bytes(m.Timestamp)
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatal(err)
		}
		err = bucket.Put(key, jsonBytes)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
}

func TestGetLatestDatabase(t *testing.T) {
	dataDir := "test"
	roomname := "testroom"

	//n := time.Now()
	//log.Println(n.Format(time.RFC3339))
	//log.Fatalln(n.UTC().Format(time.RFC3339))

	db, err := GetLatestDatabase(dataDir, roomname)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
