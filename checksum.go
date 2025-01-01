package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

var keyLength int = 16

// Generates a security key
func generateSecurityKey() []byte {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	return key
}

// Calculates an HMAC-SHA256 checksum for given data
func calculateChecksum(data []byte) string {
	h := hmac.New(sha256.New, generateSecurityKey())
	h.Write(data)
	signature := h.Sum(nil)

	return strings.ToUpper(
		fmt.Sprintf(
			"G%s",
			hex.EncodeToString(signature)[:32],
		),
	)
}
