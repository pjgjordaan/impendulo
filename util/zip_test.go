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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		if !bytes.Equal(files[zf.Name], extracted) {
			t.Error("Expected %s for %s but got %s.",
				string(files[zf.Name]), zf.Name, string(extracted))
		}
	} /*
		zf := new(zip.File)
		_, err = ExtractBytes(zf)
		if err == nil {
			t.Error("Expected error for empty zip file.")
		}*/
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
	for k1, v1 := range badFiles {
		found := false
		for k2, v2 := range unzipped {
			if strings.HasSuffix(k1, k2) && bytes.Equal(v1, v2) {
				found = true
				break
			}
		}
		if !found {
			t.Error(fmt.Errorf("Zip error, could not find value %s for %s.",
				string(v1), k1))
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
	/*badFiles := []*zip.File{new(zip.File), nil}
	for _, zf := range badFiles {
		err = ExtractFile(zf, dir)
		if err == nil {
			t.Error("Expected error for invalid zip file.")
		}
	}*/
	badLocations := []string{"/dev", "/gibberish"}
	for _, loc := range badLocations {
		err = ExtractFile(zr.File[0], loc)
		if err == nil {
			t.Errorf("Expected error for invalid directory %s.", loc)
		}
	}
}
