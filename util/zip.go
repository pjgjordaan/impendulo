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
		err = fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
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
	frc, err := zf.Open()
	if err != nil {
		return fmt.Errorf("Encountered error %q while opening zip file %q", err, zf)
	}
	defer frc.Close()
	path := filepath.Join(dir, zf.Name)
	if zf.FileInfo().IsDir() {
		err = os.MkdirAll(path, DPERM)
		if err != nil {
			return fmt.Errorf("Encountered error %q while creating directory %q", err, path)
		}
	} else {
		err = os.MkdirAll(filepath.Dir(path), DPERM)
		if err != nil {
			return fmt.Errorf("Encountered error %q while creating directory %q", err, path)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return fmt.Errorf("Encountered error %q while opening file %q", err, path)
		}
		defer f.Close()
		_, err = io.Copy(f, frc)
		if err != nil {
			return fmt.Errorf("Encountered error %q while copying from %q to %q", err, frc, f)
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
		err = fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
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
		err = fmt.Errorf("Encountered error %q while opening zip file %q", err, zf)
	} else{
		defer frc.Close()
		ret = ReadBytes(frc)
	}
	return
}

//Zip creates a zip archive from a map which has file names as its keys and file contents as its values.
func Zip(files map[string][]byte) (ret []byte, err error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, data := range files {
		err = AddToZip(zw, name, data)
		if err != nil{
			return
		}
	}
	err = zw.Close()
	if err != nil {
		err = fmt.Errorf("Encountered error %q while closing zip %q", err, zw)
	} else{
		ret = buf.Bytes()
	}
	return
}

func AddToZip(zw *zip.Writer, name string, data []byte)(err error){
	f, err := zw.Create(name)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while creating file %q in zip %q", err, name, zw)
		return
	}
	_, err = f.Write(data)
	if err != nil {
		err = fmt.Errorf("Encountered error %q while writing to file %q in zip %q", err, f, zw)
	}
	return
}
