//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
	destPath, err := filepath.Rel(this.src, path)
	if err != nil {
		//Should never happen but lets still handle it.
		return
	}
	if f == nil {
		return
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
