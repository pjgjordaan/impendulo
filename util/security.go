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
func CookieKeys() (auth, enc []byte, err error) {
	auth, err = cookieKey(authName)
	if err != nil {
		return
	}
	enc, err = cookieKey(encName)
	return
}

func cookieKey(fname string) (data []byte, err error) {
	base, err := BaseDir()
	if err != nil {
		return
	}
	fpath := filepath.Join(base, fname)
	data, rerr := ioutil.ReadFile(fpath)
	if rerr == nil {
		return
	}
	data = securecookie.GenerateRandomKey(32)
	err = SaveFile(fpath, data)
	return
}

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
