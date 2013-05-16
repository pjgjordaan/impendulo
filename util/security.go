package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
)

//Validate authenticates a provided password against a hashed password.
func Validate(hashed, salt, pword string) bool {
	computed := ComputeHash(pword, salt)
	return hashed == computed
}

//Hash hashes the provided password.
func Hash(pword string) (hash, salt string) {
	salt = GenString(32)
	return ComputeHash(pword, salt), salt
}

//ComputeHash computes the hash for a password and its salt.
func ComputeHash(pword, salt string) string {
	h := sha1.New()
	io.WriteString(h, pword+salt)
	return hex.EncodeToString(h.Sum(nil))
}

//GenString generates a psuedo-random string using the crypto/rand package.
func GenString(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}
