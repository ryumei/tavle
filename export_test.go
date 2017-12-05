package main

import (
	"os"
	"path"
	"testing"
)

func TestDectateCSV(t *testing.T) {
	msg := Message{
		Room: "testroom",
	}
	parentDir := "test/export/TestDectateCSV"
	prepareLogDir(path.Join(parentDir, "chatlog.csv"))
	if err := dectateCSV(msg, parentDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll("test/export"); err != nil {
			t.Fatal(err)
		}
	}()
}

func TestDectateCSVFailure(t *testing.T) {
	msg := Message{
		Room: "testroom",
	}
	parentDir := "test/export/TestDectateCSV/notexist"
	if err := dectateCSV(msg, parentDir); err == nil {
		t.Fatal("")
	}
	defer func() {
		if err := os.RemoveAll("test/export"); err != nil {
			t.Fatal(err)
		}
	}()
}
