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
