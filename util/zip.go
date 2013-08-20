package util

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"fmt"
)

//Unzip extracts a file (given as a []byte) to dir.
func Unzip(dir string, data []byte) (err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		err = &IOError{data, "creating zip reader from", err}
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
	if zf == nil{
		return fmt.Errorf("Could not extract nil *zip.File")
	}
	frc, err := zf.Open()
	if err != nil {
		return &IOError{zf, "opening zipfile", err}
	}
	defer frc.Close()
	path := filepath.Join(dir, zf.Name)
	if zf.FileInfo().IsDir() {
		err = os.MkdirAll(path, DPERM)
		if err != nil {
			return &IOError{path, "creating directory", err}
		}
	} else {
		err = os.MkdirAll(filepath.Dir(path), DPERM)
		if err != nil {
			return &IOError{path, "creating directory", err}
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return &IOError{path, "opening", err}
		}
		defer f.Close()
		_, err = io.Copy(f, frc)
		if err != nil {
			return &IOError{f, "copying to", err}
		}
	}
	return nil
}

//UnzipToMap reads the contents of a zip file into a map.
//Each file's path is a map key and its data is the associated value.
func UnzipToMap(data []byte) (ret map[string][]byte, err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		err = &IOError{data, "creating zip reader from", err}
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
		err = &IOError{zf, "opening zipfile", err}
	} else {
		defer frc.Close()
		ret = ReadBytes(frc)
	}
	return
}

//Zip creates a zip archive from a map which has file names as its keys and
//file contents as its values.
func Zip(files map[string][]byte) (ret []byte, err error) {
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
		err = &IOError{zw, "closing zip", err}
	} else {
		ret = buf.Bytes()
	}
	return
}

//AddToZip adds a new file to a zip.Writer.
func AddToZip(zw *zip.Writer, name string, data []byte) (err error) {
	f, err := zw.Create(name)
	if err != nil {
		err = &IOError{name, "creating zipfile", err}
		return
	}
	_, err = f.Write(data)
	if err != nil {
		err = &IOError{f, "writing to zipfile", err}
	}
	return
}
