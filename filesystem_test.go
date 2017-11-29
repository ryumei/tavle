package main

import (
	"log"
	"os"
	"path"
	"testing"
)

// TestPrepareLogDir ディレクトリの作成
func TestPrepareLogDir(t *testing.T) {
	if err := prepareLogDir("./test/filesystem_test/TestPrepareLogDir/test.txt"); err != nil {
		t.Fatalf("Failed to create nested . %v", err)
	}
	defer func() {
		if err := os.RemoveAll("./test/filesystem_test"); err != nil {
			log.Printf("[ERROR] %v", err)
		}
	}()
}

// TestPrepareLogDirFailure ディレクトリの作成エラー
func TestPrepareLogDirFailure(t *testing.T) {
	filename := "test_file.txt"
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(filename)
	}()

	invalidPath := path.Join(filename, "test_dir", "test.txt")
	if err := prepareLogDir(invalidPath); err == nil {
		t.Fatalf("Failed to create dir invalid path. %v", err)
	}
}
