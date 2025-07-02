package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func StrHash(s1 string) string {
	hash := sha256.New()
	hash.Write([]byte(s1))
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes)
}
