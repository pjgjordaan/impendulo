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

//Package util provides utility methods for performing operations which are used
//throughout impendulo such as io, type conversion and logging.
package util

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/godfried/impendulo/util/errors"

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	DPERM = 0777
	FPERM = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	EOT   = "\u0004"
)

type (
	//copier
	copier struct {
		dest, src string
	}

	finder struct {
		l, n string
	}
)

//ReadData reads data from a reader until io.EOF or []byte("eof") is encountered.
func ReadData(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("Can't read from a nil io.Reader")
	}
	b := new(bytes.Buffer)
	t := []byte(EOT)
	p := make([]byte, 2048)
	busy := true
	for busy {
		n, e := r.Read(p)
		s := p[:n]
		if e == io.EOF {
			busy = false
		} else if e != nil {
			return nil, errors.NewUtil(r, "reading from", e)
		} else if bytes.HasSuffix(s, t) {
			s = s[:len(s)-len(t)]
			busy = false
		}
		b.Write(s)
	}
	return b.Bytes(), nil
}

func SaveTemp(d []byte) (string, error) {
	f, e := ioutil.TempFile("", "")
	if e != nil {
		return "", e
	}
	if _, e = f.Write(d); e != nil {
		return "", e
	}
	return f.Name(), nil
}

//SaveFile saves a file (given as a []byte) as n.
func SaveFile(n string, d []byte) error {
	if e := os.MkdirAll(filepath.Dir(n), DPERM); e != nil {
		return errors.NewUtil(n, "creating", e)
	}
	f, e := os.Create(n)
	if e != nil {
		return errors.NewUtil(n, "creating", e)
	}
	if _, e = f.Write(d); e != nil {
		return errors.NewUtil(n, "writing to", e)
	}
	return nil
}

//ReadBytes reads bytes from a reader until io.EOF is encountered.
//If the reader can't be read an empty []byte is returned.
func ReadBytes(r io.Reader) []byte {
	if r == nil {
		return make([]byte, 0)
	}
	b := new(bytes.Buffer)
	if _, e := b.ReadFrom(r); e != nil {
		return make([]byte, 0)
	}
	return b.Bytes()
}

//GetPackage retrieves the package name from a Java source file.
func GetPackage(r io.Reader) string {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		if s.Text() == "package" {
			s.Scan()
			return strings.Split(s.Text(), ";")[0]
		}
	}
	return ""
}

//copyFile
func (c *copier) copyFile(p string, f os.FileInfo, e error) error {
	if e != nil {
		return e
	}
	return c.copy(p, f)
}

//copy
func (c *copier) copy(p string, f os.FileInfo) error {
	if f == nil {
		return nil
	}
	dp, e := filepath.Rel(c.src, p)
	if e != nil {
		return e
	}
	if dp == "." {
		dp = ""
	}
	dp = filepath.Join(c.dest, dp)
	if f.IsDir() {
		return os.MkdirAll(dp, os.ModePerm)
	}
	sf, e := os.Open(p)
	if e != nil {
		return e
	}
	df, e := os.Create(dp)
	if e != nil {
		return e
	}
	_, e = io.Copy(df, sf)
	return e
}

//Copy copies the contents of s to d.
func Copy(d, s string) error {
	c := &copier{d, s}
	return filepath.Walk(s, c.copyFile)
}

//Exists checks whether a given path exists.
func Exists(p string) bool {
	_, e := os.Stat(p)
	return e == nil
}

//IsFile checks whether a given path is a file.
func IsFile(p string) bool {
	s, e := os.Stat(p)
	if e != nil {
		return false
	}
	return !s.IsDir()
}

//IsDir checks whether a given path is a directory.
func IsDir(p string) bool {
	s, e := os.Stat(p)
	if e != nil {
		return false
	}
	return s.IsDir()
}

//IsExec checks whether a given path is an executable file.
func IsExec(p string) bool {
	s, e := os.Stat(p)
	if e != nil {
		return false
	}
	return !s.IsDir() && (s.Mode()&0100) != 0
}

func CopyFile(d, s string) error {
	r, e := os.Open(s)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(d), DPERM); e != nil {
		return e
	}
	w, e := os.Create(d)
	if e != nil {
		return e
	}
	_, e = io.Copy(w, r)
	return e
}

func Extension(n string) (string, string) {
	s := strings.Split(n, ".")
	switch len(s) {
	case 0, 1:
		return n, ""
	default:
		return strings.Join(s[0:len(s)-1], "."), s[len(s)-1]
	}
}

func (f *finder) walk(p string, i os.FileInfo, e error) error {
	if e != nil {
		return e
	}
	if i.IsDir() && strings.HasSuffix(p, f.n) {
		f.l = p
		return errors.Found
	}
	return nil
}

func LocateDirectory(src, name string) (string, error) {
	f := &finder{n: name}
	if e := filepath.Walk(src, f.walk); e != nil && e != errors.Found {
		return "", e
	}
	return f.l, nil
}
