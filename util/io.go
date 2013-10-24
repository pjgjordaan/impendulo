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
	"io"
	"os"
	"os/user"
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
)

//BaseDir retrieves the Impendulo directory.
func BaseDir() string {
	cur, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(cur.HomeDir, ".impendulo")
}

//ReadData reads data from a reader until io.EOF or []byte("eof") is encountered.
func ReadData(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("Can't read from a nil io.Reader")
	}
	buffer := new(bytes.Buffer)
	eot := []byte(EOT)
	p := make([]byte, 2048)
	busy := true
	for busy {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if err == io.EOF {
			busy = false
		} else if err != nil {
			return nil, &UtilError{r, "reading from", err}
		} else if bytes.HasSuffix(read, eot) {
			read = read[:len(read)-len(eot)]
			busy = false
		}
		buffer.Write(read)
	}
	return buffer.Bytes(), nil
}

//SaveFile saves a file (given as a []byte) as fname.
func SaveFile(fname string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(fname), DPERM)
	if err != nil {
		return &UtilError{fname, "creating", err}
	}
	f, err := os.Create(fname)
	if err != nil {
		return &UtilError{fname, "creating", err}
	}
	_, err = f.Write(data)
	if err != nil {
		return &UtilError{fname, "writing to", err}
	}
	return nil
}

//ReadBytes reads bytes from a reader until io.EOF is encountered.
//If the reader can't be read an empty []byte is returned.
func ReadBytes(r io.Reader) []byte {
	if r == nil {
		return make([]byte, 0)
	}
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return make([]byte, 0)
	} else {
		return buffer.Bytes()
	}
}

//GetPackage retrieves the package name from a Java source file.
func GetPackage(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "package" {
			scanner.Scan()
			return strings.Split(scanner.Text(), ";")[0]
		}
	}
	return ""
}

//copyFile
func (this *copier) copyFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	return this.copy(path, f)
}

//copy
func (this *copier) copy(path string, f os.FileInfo) (err error) {
	if err != nil {
		return
	}
	if f == nil {
		return
	}
	destPath, err := filepath.Rel(this.src, path)
	if err != nil {
		return
	}
	if destPath == "." {
		destPath = ""
	}
	destPath = filepath.Join(this.dest, destPath)
	if f.IsDir() {
		err = os.MkdirAll(destPath, os.ModePerm)
		return
	}
	srcFile, err := os.Open(path)
	if err != nil {
		return
	}
	destFile, err := os.Create(destPath)
	if err != nil {
		return
	}
	_, err = io.Copy(destFile, srcFile)
	return
}

//Copy copies the contents of src to dest.
func Copy(dest, src string) error {
	c := &copier{dest, src}
	return filepath.Walk(src, c.copyFile)
}

//Exists checks whether a given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

//IsFile checks whether a given path is a file.
func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

//IsDir checks whether a given path is a directory.
func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

//IsExec checks whether a given path is an executable file.
func IsExec(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir() && (stat.Mode()&0100) != 0
}

func CopyFile(dest, src string) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return
	}
	w, err := os.Create(dest)
	if err != nil {
		return
	}
	_, err = io.Copy(w, r)
	return
}
