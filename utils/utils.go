package utils

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"intlola/client"
	"io/ioutil"
	"os"
"strings"
"io"
)

const SEP = string(os.PathSeparator)
const PERM = 0777
const BASE_DIR = "data"

func WriteFile(file string, data *bytes.Buffer) error {
	Log("Writing to: ", file)
	err := ioutil.WriteFile(BASE_DIR+SEP+file, data.Bytes(), 0666)
	return err
}

func ReadFile(fname string)([]byte, error){
	return ioutil.ReadFile(BASE_DIR+SEP+fname)
}

func ReadUsers(fname string)(map[string] string, error){
	users := make(map[string] string)
	data, err := ioutil.ReadFile(fname)
	buff := bytes.NewBuffer(data)
	line, err := buff.ReadString(byte('\n'))
	for err == nil{
		vals := strings.Split(line, ":")
		users[strings.TrimSpace(vals[0])] = strings.TrimSpace(vals[1])
		line, err = buff.ReadString(byte('\n'))
	}
	if err == io.EOF{
		err = nil
	}
	return users, err
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}

func MkDir(dir string) (err error) {
	return os.Mkdir(BASE_DIR+SEP+dir, PERM)
}

func CreateUserProject(c *client.Client) error {
	MkDir(c.Project)
	return MkDir(c.Project+SEP+c.Name)
}

func ZipDir(dir, fname string) (err error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	finfos, err := ioutil.ReadDir(BASE_DIR+SEP+dir)
	if err == nil {
		for _, file := range finfos {
			if !file.IsDir() {
				f, err := w.Create(file.Name())
				if err != nil {
					break
				}
				contents, err := ReadFile(dir + SEP + file.Name())
				if err != nil {
					break
				}
				_, err = f.Write(contents)
				if err != nil {
					break
				}
			}
		}

	}
	w.Close()
	if err == nil {
		err = WriteFile(dir+SEP+fname, buf)
	}
	Log(err)
	return err
}

func JSONValue(jobj map[string]interface{}, key string) (string, error) {
	ival, err := jobj[key]
	if !err {
		return "", errors.New(key + " not found in JSON Object")
	}
	switch val := ival.(type) {
	case string:
		return val, nil
	}
	return "", errors.New("Invalid type in JSON parameter")
}
