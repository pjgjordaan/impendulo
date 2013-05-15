package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
)

func Validate(hashed, salt, pword string) bool {
	computed := computeHash(pword, salt)
	return hashed == computed
}

func Hash(pword string) (hash, salt string) {
	salt = GenString(32)
	return computeHash(pword, salt), salt
}

func computeHash(pword, salt string) string {
	h := sha1.New()
	io.WriteString(h, pword+salt)
	return hex.EncodeToString(h.Sum(nil))
}

func GenString(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}
