package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
)

/*
type manager struct {
	dbs map[string]*bolt.DB
}

var Mgr *manager

func init() {
	Mgr = &manager{dbs: map[string]*bolt.DB{}}
}
*/

// BucketFormat is Datetime format for bucket name
var BucketFormat = "2006.01.02"

/*
 * DB structure
 * DB file: roomname
 *   - bucket: based on DATE
 *     - key: timestamp + SHA1(username)
 * e.g.)
 *   data/foyer.db -> 2018.01.25 -> { hash("timestamp+username"): json(Message) }
 * housekeeping -> keep 2 month events in database
 */

func GetReadOnlyDB(dataDir string, roomname string) (*bolt.DB, error) {
	return getDatabase(dataDir, roomname, 0644)
}

func GetWritableDB(dataDir string, roomname string) (*bolt.DB, error) {
	return getDatabase(dataDir, roomname, 0644)
}

/*
 *
 * [NOTE] db must be closed with defer.
 */
func getDatabase(dataDir string, roomname string, permission os.FileMode) (*bolt.DB, error) {
	// Based on localtime
	dbFile := fmt.Sprintf("%s.db", roomname)

	_, err := os.Stat(dataDir)
	if err != nil {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			log.Fatalf("Failed to created data dir %s. %v", dataDir, err)
		}
	}
	dbPath := filepath.Join(dataDir, dbFile)
	db, err := bolt.Open(dbPath, permission, nil)
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
	binary.BigEndian.PutUint64(b, uint64(ns))
	return b
}

func bytes2time(b []byte) time.Time {
	ns := int64(binary.BigEndian.Uint64(b))
	return time.Unix(0, ns)
}

func dbKeySHA1(t time.Time, user string) []byte {
	var k = make([]byte, 0, 8+sha1.Size)
	for _, c := range time2bytes(t) {
		k = append(k, c)
	}
	for _, c := range sha1.Sum([]byte(user)) {
		k = append(k, c)
	}
	return k
}

// timestamp bytes only hash key padded with zero
func dbBoundKeySHA1(t time.Time) []byte {
	var k = make([]byte, 0, 8+sha1.Size)
	for _, c := range time2bytes(t) {
		k = append(k, c)
	}
	// pad trailing zeros
	for i := sha1.Size; i > 0; i-- {
		k = append(k, 0)
	}

	return k
}

// SavePost メッセージを DB に保管します。
func SavePost(m Message, dataDir string, secret []byte) error {
	db, err := GetWritableDB(dataDir, m.Room)
	if err != nil {
		log.Print("[ERROR] Failed to open the DB", err)
		return nil
	}
	defer db.Close()

	bucketName := m.Timestamp.UTC().Format(BucketFormat)
	key := dbKeySHA1(m.Timestamp, m.Username)
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		log.Print("[ERROR] Failed to JSONize", err)
		return nil
	}
	encrypted, err := Encrypt(jsonBytes, secret)
	if err != nil {
		log.Print("[ERROR] Failed to encrypt a message", err)
		return nil
	}

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Print("[ERROR] Failed to get the bucket", err)
			return nil
		}
		err = bucket.Put(key, encrypted)
		if err != nil {
			log.Print("[ERROR] Failed to save a message", err)
			return nil
		}
		return nil
	})

	return nil
}

func LoadPosts(room string, latest time.Time, durationSec int, dataDir string, secret []byte) ([]Message, error) {
	db, err := GetReadOnlyDB(dataDir, room)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	latest = latest.UTC()
	earliest := latest.Add(-time.Duration(durationSec) * time.Second)

	// Earliest and latest bound
	min := dbBoundKeySHA1(earliest)
	max := dbBoundKeySHA1(latest.Add(1 + time.Nanosecond))

	aDay := 24 * time.Hour

	var messages []Message
	for cursor := earliest; latest.Sub(cursor)+aDay > 0; cursor = cursor.AddDate(0, 0, 1) {

		bucketName := cursor.UTC().Format(BucketFormat)
		db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			if bucket == nil {
				return nil // Skip. no bucket found
			}

			c := bucket.Cursor()
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) < 0; k, v = c.Next() {

				decrypted, err := Decrypt(v, secret)
				if err != nil {
					log.Fatalf("Decryption error: %v ", err)
				}
				var msg Message
				if err := json.Unmarshal(decrypted, &msg); err != nil {
					log.Println("[WARN] JSON Unmarshal error:", err)
					continue
				}
				messages = append(messages, msg)
			}

			return nil
		})
	}
	return messages, nil
}
