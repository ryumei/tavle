package main

import "testing"

func TestPrepareLogDir(t *testing.T) {
	if err := prepareLogDir("./test/filesystem_test/TestPrepareLogDir/test.txt"); err != nil {
		t.Fatalf("Failed to create nested . %v", err)
	}
}

func TestPrepareLogDirFailure(t *testing.T) {
	if err := prepareLogDir("/not_exist_dir/test/filesystem_test/TestPrepareLogDir/test.txt"); err == nil {
		t.Fatalf("Failed to create nested . %v", err)
	}
}
