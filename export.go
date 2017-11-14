package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

// CSV ファイルへチャットログを書き出す
func dectateCSV(msg Message) {
	fname := fmt.Sprintf("%s-%s.csv", msg.Room, time.Now().Format("20060102"))
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("[ERROR] failed to write file %v", err)
		return
	}
	defer f.Close()
	writer := csv.NewWriter(f)

	row := []string{
		msg.Timestamp.Format(time.RFC3339),
		msg.Room,
		msg.Username,
		msg.Email,
		msg.Message,
	}

	if err := writer.Write(row); err != nil {
		log.Printf("[ERROR] failed to write file %v", err)
		return
	}
	writer.Flush()
}
