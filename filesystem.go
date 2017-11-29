package main

import (
	"log"
	"os"
	"path"
)

func prepareLogDir(logpath string) error {
	parent := path.Dir(logpath)
	if err := os.MkdirAll(parent, os.ModePerm); err != nil {
		log.Printf("[ERROR] %v", err)
		return err
	}
	return nil
}
