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
	"io"
	"os"
	"path/filepath"
)

//Unzip extracts a file (given as a []byte) to dir.
func Unzip(dir string, data []byte) (err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		err = &UtilError{data, "creating zip reader from", err}
		return
	}
	for _, zf := range zr.File {
		err = ExtractFile(zf, dir)
		if err != nil {
			return
		}
	}
	return
}

//ExtractFile extracts the contents of a zip.File and saves it in dir.
func ExtractFile(zf *zip.File, dir string) error {
	if zf == nil {
		return fmt.Errorf("Could not extract nil *zip.File")
	}
	frc, err := zf.Open()
	if err != nil {
		return &UtilError{zf, "opening zipfile", err}
	}
	defer frc.Close()
	path := filepath.Join(dir, zf.Name)
	if zf.FileInfo().IsDir() {
		err = os.MkdirAll(path, DPERM)
		if err != nil {
			err = &UtilError{path, "creating directory", err}
		}
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), DPERM)
	if err != nil {
		return &UtilError{path, "creating directory", err}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
	if err != nil {
		return &UtilError{path, "opening", err}
	}
	defer f.Close()
	_, err = io.Copy(f, frc)
	if err != nil {
		err = &UtilError{f, "copying to", err}
	}
	return err
}

//UnzipToMap reads the contents of a zip file into a map.
//Each file's path is a map key and its data is the associated value.
func UnzipToMap(data []byte) (ret map[string][]byte, err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		err = &UtilError{data, "creating zip reader from", err}
		return
	}
	ret = make(map[string][]byte)
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		ret[zf.FileInfo().Name()], err = ExtractBytes(zf)
		if err != nil {
			return
		}
	}
	return
}

//ExtractBytes extracts data from a zip.File.
func ExtractBytes(zf *zip.File) (ret []byte, err error) {
	frc, err := zf.Open()
	if err != nil {
		err = &UtilError{zf, "opening zipfile", err}
	} else {
		defer frc.Close()
		ret = ReadBytes(frc)
	}
	return
}

//Zip creates a zip archive from a map which has file names as its keys and
//file contents as its values.
func ZipMap(files map[string][]byte) (ret []byte, err error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, data := range files {
		err = AddToZip(zw, name, data)
		if err != nil {
			return
		}
	}
	err = zw.Close()
	if err != nil {
		err = &UtilError{zw, "closing zip", err}
	} else {
		ret = buf.Bytes()
	}
	return
}

//AddToZip adds a new file to a zip.Writer.
func AddToZip(zw *zip.Writer, name string, data []byte) (err error) {
	f, err := zw.Create(name)
	if err != nil {
		err = &UtilError{name, "creating zipfile", err}
		return
	}
	_, err = f.Write(data)
	if err != nil {
		err = &UtilError{f, "writing to zipfile", err}
	}
	return
}
