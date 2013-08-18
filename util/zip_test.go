package util

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"archive/zip"
)

func TestZip(t *testing.T) {
	goodFiles := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."), 
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"), 
		"todo.txt": []byte("Get animal handling licence.\nWrite more examples."),
	}
	_, err := Zip(goodFiles)
	if err != nil {
		t.Error(err)
	}
}


func TestExtractBytes(t *testing.T){
	zr, err := zip.OpenReader("testArchive.zip")
	if err != nil {
		t.Error(err)
	}
	for _, zf := range zr.File{
		_, err = ExtractBytes(zf)
		if err != nil{
			t.Error(err)
		}
	}
	zf := new(zip.File)
	_, err = ExtractBytes(zf)
	if err == nil{
		t.Error("Expected error for empty zip file.")
	}
}

func TestUnzipToMap(t *testing.T) {
	goodFiles := map[string][]byte{
		"readme.txt": []byte("This archive contains some text files."), 
		"gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"), 
		"todo.txt": []byte("Get animal handling licence.\nWrite more examples."),
	}
	zipped, err := Zip(goodFiles)
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
		"\\hi": nil,
	}
	zipped, err = Zip(badFiles)
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


func TestUnzip(t *testing.T){
	f, err := os.Open("testArchive.zip")
	if err != nil {
		t.Error(err)
	}
	buff := new(bytes.Buffer)
	_, err = buff.ReadFrom(f)
	if err != nil {
		t.Error(err)
	}
	os.MkdirAll("/tmp/unzipped", os.ModeDir | os.ModePerm)
	defer os.Remove("/tmp/unzipped")
	err = Unzip("/tmp/unzipped", buff.Bytes())
	if err != nil{
		t.Error(err)
	}
}
