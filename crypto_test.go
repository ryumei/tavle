package main

import (
	"testing"
)

func TestEncrypt(t *testing.T) {
	for _, item := range encryptionTests {
		data := item.data
		secret := []byte(item.secret)

		encrypted, err := encrypt(data, secret)
		if err != nil {
			t.Fatalf("%v", err)
		}

		decrypted, err := decrypt(encrypted, secret)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if data != decrypted {
			t.Fatalf("Failed ecryption/decryption %v %v", data, decrypted)
		}
	}
}
