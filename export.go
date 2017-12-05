package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

// CSV ファイルへチャットログを書き出す
func dectateCSV(msg Message, parent string) error {
	// NOTE サーバ側の timezone がそれぞれで異なると、
	// 同時刻でも別の日付ログファイルに入る可能性あり。
	// マージ後に考えれば良いか。
	// キャッシュ層を入れれば、そこで吸収できるはず。
	fpath := path.Join(parent,
		fmt.Sprintf("%s-%s.csv", msg.Room, time.Now().Format("20060102")))

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("[ERROR] failed to write file %v", err)
		return err
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
		return err
	}
	writer.Flush()
	return nil
}
