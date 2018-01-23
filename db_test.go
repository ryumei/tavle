package main

import (
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
 *     - key: timestamp
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

func WritePost(m Message) {
	dataDir := "."
	db, err := GetLatestDatabase(dataDir, m.Room)
	if err != nil {
		log.Fatal(err)
	}
	bucketName := m.Timestamp.Format("2006.01.02")
	key := m.Timestamp.UnmarshalBinary()
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatal(err)
		}
		err = bucket.Put([]byte(key), jsonBytes)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

}

func TestGetLatestDatabase(t *testing.T) {
	dataDir := "test"
	roomname := "testroom"

	db, err := GetLatestDatabase(dataDir, roomname)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
