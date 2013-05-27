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
func Unzip(dir string, data []byte) error {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
	}
	for _, zf := range zr.File {
		err = ExtractFile(zf, dir)
		if err != nil {
			return err
		}
	}
	return nil
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
func UnzipToMap(data []byte) (map[string][]byte, error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
	}
	extracted := make(map[string][]byte)
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		data, err := ExtractBytes(zf)
		if err != nil {
			return nil, err
		}
		extracted[zf.FileInfo().Name()] = data
	}
	return extracted, nil
}

//ExtractBytes extracts data from a zip.File. 
func ExtractBytes(zf *zip.File) ([]byte, error) {
	frc, err := zf.Open()
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while opening zip file %q", err, zf)
	}
	defer frc.Close()
	return ReadBytes(frc), nil
}

//Zip creates a zip archive from a map which has file names as its keys and file contents as its values.
func Zip(files map[string][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, fmt.Errorf("Encountered error %q while creating file %q in zip %q", err, name, w)
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, fmt.Errorf("Encountered error %q while writing to file %q in zip %q", err, f, w)
		}
	}
	err := w.Close()
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while closing zip %q", err, w)
	}
	return buf.Bytes(), nil
}
