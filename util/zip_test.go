package util

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"testing"
	"path/filepath"
)

func TestZipMap(t *testing.T) {
	goodFiles := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
	}
	_, err := ZipMap(goodFiles)
	if err != nil {
		t.Error(err)
	}
}

func TestExtractBytes(t *testing.T) {
	files := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
	}
	zipped, err := ZipMap(files)
	r := bytes.NewReader(zipped)
	zr, err := zip.NewReader(r, int64(r.Len()))
	if err != nil {
		t.Error(err)
	}
	for _, zf := range zr.File {
		extracted, err := ExtractBytes(zf)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(files[zf.Name], extracted){
			t.Error("Expected %s for %s but got %s.",
				string(files[zf.Name]), zf.Name, string(extracted))
		} 
	}
	zf := new(zip.File)
	_, err = ExtractBytes(zf)
	if err == nil {
		t.Error("Expected error for empty zip file.")
	}
}

func TestUnzipToMap(t *testing.T) {
	goodFiles := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
	}
	zipped, err := ZipMap(goodFiles)
	if err != nil {
		t.Error(err)
	}
	unzipped, err := UnzipToMap(zipped)
	if len(goodFiles) != len(unzipped) {
		t.Error(errors.New("Zip error; invalid size"))
	}
	for k, v := range goodFiles {
		if !bytes.Equal(v, unzipped[k]) {
			t.Error(errors.New("Zip error, values not equal."))
		}
	}
	badFiles := map[string][]byte{
		"C://hi.txt": []byte("This archive contains some text files."),
		"/root/sudo": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"\\hi":       nil,
	}
	zipped, err = ZipMap(badFiles)
	if err != nil {
		t.Error(err)
	}
	unzipped, err = UnzipToMap(zipped)
	if err != nil {
		t.Error(err)
	}
	if len(badFiles) != len(unzipped) {
		t.Error(errors.New("Zip error; invalid size"))
	}
	for k, v := range badFiles {
		if !bytes.Equal(v, unzipped[k]) {
			t.Error(errors.New("Zip error, values not equal."))
		}
	}

	_, err = UnzipToMap(nil)
	if err == nil {
		t.Error("Expected error.")
	}
}

func TestUnzip(t *testing.T) {
	files := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
	}
	zipped, err := ZipMap(files)
	buff := bytes.NewBuffer(zipped)
	dir := filepath.Join(os.TempDir(), "unzipped")
	os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	defer os.RemoveAll(dir)
	err = Unzip(dir, buff.Bytes())
	if err != nil {
		t.Error(err)
	}
}

func TestExtractFile(t *testing.T) {
	files := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."),
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"),
		"todo.txt":   []byte("Get animal handling licence.\nWrite more examples."),
	}
	zipped, err := ZipMap(files)
	r := bytes.NewReader(zipped)
	zr, err := zip.NewReader(r, int64(r.Len()))
	if err != nil {
		t.Error(err)
	}
	dir := filepath.Join(os.TempDir(), "extractFile")
	os.MkdirAll(dir, os.ModeDir|os.ModePerm)
	defer os.RemoveAll(dir)
	for _, zf := range zr.File {
		err := ExtractFile(zf, dir)
		if err != nil {
			t.Error(err)
		} 
	}
	badFiles := []*zip.File{new(zip.File), nil} 
	for _, zf := range badFiles{
		err = ExtractFile(zf, dir)
		if err == nil {
			t.Error("Expected error for invalid zip file.")
		}
	}
	badLocations := []string{ "/dev", "/gibberish"}
	for _, loc := range badLocations{
		err = ExtractFile(zr.File[0], loc)
		if err == nil {
			t.Errorf("Expected error for invalid directory %s.", loc)
		}
	}
}
