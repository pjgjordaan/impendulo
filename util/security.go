//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package util

import (
	"code.google.com/p/gorilla/securecookie"

	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
)

var (
	authName = "authentication.key"
	encName  = "encryption.key"
)

//CookieKeys generates cookie keys.
func CookieKeys() ([]byte, []byte, error) {
	a, e := cookieKey(authName)
	if e != nil {
		return nil, nil, e
	}
	enc, e := cookieKey(encName)
	if e != nil {
		return nil, nil, e
	}
	return a, enc, nil
}

func cookieKey(n string) ([]byte, error) {
	b, e := BaseDir()
	if e != nil {
		return nil, e
	}
	p := filepath.Join(b, n)
	d, e := ioutil.ReadFile(p)
	if e == nil {
		return d, nil
	}
	d = securecookie.GenerateRandomKey(32)
	if e = SaveFile(p, d); e != nil {
		return nil, e
	}
	return d, nil
}

//Validate authenticates a provided password against a hashed password.
func Validate(h, s, p string) bool {
	return h == ComputeHash(p, s)
}

//Hash hashes the provided password and returns the hash as well as the salt used.
func Hash(p string) (string, string) {
	s := GenString(32)
	return ComputeHash(p, s), s
}

//ComputeHash computes the hash for a password and its salt.
func ComputeHash(p, s string) string {
	h := sha1.New()
	io.WriteString(h, p+s)
	return hex.EncodeToString(h.Sum(nil))
}

//GenString generates a psuedo-random string using the crypto/rand package.
func GenString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	e := base64.StdEncoding
	d := make([]byte, e.EncodedLen(len(b)))
	e.Encode(d, b)
	return string(d)
}
