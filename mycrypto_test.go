package main

import (
	"bytes"
	"testing"
)

func TestEncrypt(t *testing.T) {
	for _, item := range encryptionTests {
		data := []byte(item.data)
		secret := []byte(item.secret)

		encrypted, err := Encrypt(data, secret)
		if err != nil {
			t.Fatalf("%v", err)
		}

		decrypted, err := Decrypt(encrypted, secret)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !bytes.Equal([]byte(data), decrypted) {
			t.Fatalf("Failed ecryption/decryption %v %v", data, decrypted)
		}
	}
}
