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
	"archive/zip"
	"bytes"
	"fmt"

	"github.com/godfried/impendulo/util/errors"

	"io"
	"os"
	"path/filepath"
)

type (
	KVWriter interface {
		Write(string, []byte) error
	}

	KVReader interface {
		Next() (string, []byte, error)
	}
)

//Unzip extracts a file (given as a []byte) to dir.
func Unzip(d string, p []byte) error {
	br := bytes.NewReader(p)
	zr, e := zip.NewReader(br, int64(br.Len()))
	if e != nil {
		return errors.NewUtil(p, "creating zip reader from", e)
	}
	for _, f := range zr.File {
		if e = ExtractFile(f, d); e != nil {
			return e
		}
	}
	return nil
}

//ExtractFile extracts the contents of a zip.File and saves it in dir.
func ExtractFile(zf *zip.File, d string) error {
	if zf == nil {
		return fmt.Errorf("Could not extract nil *zip.File")
	}
	r, e := zf.Open()
	if e != nil {
		return errors.NewUtil(zf, "opening zipfile", e)
	}
	defer r.Close()
	p := filepath.Join(d, zf.Name)
	if zf.FileInfo().IsDir() {
		if e = os.MkdirAll(p, DPERM); e != nil {
			return errors.NewUtil(p, "creating directory", e)
		}
		return nil
	}
	if e = os.MkdirAll(filepath.Dir(p), DPERM); e != nil {
		return errors.NewUtil(p, "creating directory", e)
	}
	f, e := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
	if e != nil {
		return errors.NewUtil(p, "opening", e)
	}
	defer f.Close()
	if _, e = io.Copy(f, r); e != nil {
		return errors.NewUtil(f, "copying to", e)
	}
	return nil
}

//UnzipToMap reads the contents of a zip file into a map.
//Each file's path is a map key and its data is the associated value.
func UnzipToMap(d []byte) (map[string][]byte, error) {
	br := bytes.NewReader(d)
	zr, e := zip.NewReader(br, int64(br.Len()))
	if e != nil {
		return nil, errors.NewUtil(d, "creating zip reader from", e)
	}
	m := make(map[string][]byte)
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		m[zf.Name], e = ExtractBytes(zf)
		if e != nil {
			return nil, e
		}
	}
	return m, nil
}

func UnzipKV(w KVWriter, d []byte) error {
	br := bytes.NewReader(d)
	zr, e := zip.NewReader(br, int64(br.Len()))
	if e != nil {
		return errors.NewUtil(d, "creating zip reader from", e)
	}
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		b, e := ExtractBytes(zf)
		if e != nil {
			return e
		}
		if e = w.Write(zf.Name, b); e != nil {
			return e
		}
	}
	return nil
}

func ZipKV(r KVReader) ([]byte, error) {
	b := new(bytes.Buffer)
	zw := zip.NewWriter(b)
	for {
		n, d, e := r.Next()
		if e != nil {
			break
		}
		if e := AddToZip(zw, n, d); e != nil {
			return nil, e
		}
	}
	if e := zw.Close(); e != nil {
		return nil, errors.NewUtil(zw, "closing zip", e)
	}
	return b.Bytes(), nil
}

//ExtractBytes extracts data from a zip.File.
func ExtractBytes(zf *zip.File) ([]byte, error) {
	r, e := zf.Open()
	if e != nil {
		return nil, errors.NewUtil(zf, "opening zipfile", e)
	}
	defer r.Close()
	return ReadBytes(r), nil
}

//Zip creates a zip archive from a map which has file names as its keys and
//file contents as its values.
func ZipMap(m map[string][]byte) ([]byte, error) {
	b := new(bytes.Buffer)
	zw := zip.NewWriter(b)
	for n, d := range m {
		if e := AddToZip(zw, n, d); e != nil {
			return nil, e
		}
	}
	if e := zw.Close(); e != nil {
		return nil, errors.NewUtil(zw, "closing zip", e)
	}
	return b.Bytes(), nil
}

//AddToZip adds a new file to a zip.Writer.
func AddToZip(zw *zip.Writer, n string, d []byte) error {
	f, e := zw.Create(n)
	if e != nil {
		return errors.NewUtil(n, "creating zipfile", e)
	}
	if _, e = f.Write(d); e != nil {
		return errors.NewUtil(f, "writing to zipfile", e)
	}
	return nil
}
