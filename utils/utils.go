package utils

import (
	"archive/zip"
	"bytes"
	"errors"
	"github.com/disco-volante/intlola/client"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const SEP = string(os.PathSeparator)
const DPERM = 0777
const FPERM = 0666
const DEBUG = true
const DB_PATH = "db"
const LOG_DIR = "logs"
var BASE_DIR = "Data"
var logger *log.Logger

func init() {
	cur, err := user.Current()
	if err == nil {
		temp := cur.HomeDir + SEP + ".intlola" + SEP + BASE_DIR
		BASE_DIR = ""
		MkDir(temp)
		BASE_DIR = temp
	} else {
		MkDir("")
	}
	now := time.Now()
	y, m, d := now.Date()
	dir := LOG_DIR+SEP+strconv.Itoa(y)+SEP+m.String()+SEP+strconv.Itoa(d)
	MkDir(dir)
	fo, err := os.Create(BASE_DIR + SEP + dir + SEP + time.Now().String()+".log")
	if err != nil {
		panic(err)
	}
	logger = log.New(fo, "Inlola log", log.LstdFlags)
}

func WriteFile(file string, data *bytes.Buffer) error {
	return ioutil.WriteFile(BASE_DIR+SEP+file, data.Bytes(), FPERM)
}

func ReadFile(fname string) ([]byte, error) {
	return ioutil.ReadFile(BASE_DIR + SEP + fname)
}

func ReadUsers(fname string) (users []*client.ClientData, err error) {
	data, err := ioutil.ReadFile(fname)
	if err == nil {
		buff := bytes.NewBuffer(data)
		line, err := buff.ReadString(byte('\n'))
		users = make([]*client.ClientData, 100)
		i := 0
		for err == nil {
			vals := strings.Split(line, ":")
			user := strings.TrimSpace(vals[0])
			pword := strings.TrimSpace(vals[1])
			data := &client.ClientData{user, pword, make(map[string] int)}
			if i >= len(users){
				users = append(users, data)
			}else{
				users[i] = data
			}
			i ++
			line, err = buff.ReadString(byte('\n'))
		}
		if err == io.EOF {
			err = nil
			users = users[:i-1]
		}
	}
	return users, err
}

func Log(v ...interface{}) {
	if DEBUG {
		logger.Print(v...)
	}
}

func MkDir(dir string) (err error) {
	if strings.Contains(dir, SEP) {
		dirs := strings.Split(dir, SEP)
		cur := BASE_DIR
		for _, d := range dirs {
			cur = cur + SEP + d
			err = os.Mkdir(cur, DPERM)
		}
	} else {
		err = os.Mkdir(BASE_DIR+SEP+dir, DPERM)
	}
	return err
}

func Remove(path string) error {
	return os.RemoveAll(BASE_DIR + SEP + path)
}

func ZipProject(c *client.Client) (err error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	dir := c.Project + SEP + c.Name
	finfos, err := ioutil.ReadDir(BASE_DIR + SEP + dir)
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
	errw := w.Close()
	if err == nil {
		if errw == nil {
			path := c.Project + SEP + c.Project + strconv.Itoa(c.ProjectNum) + "_" + c.Name + ".zip"
			err = WriteFile(path, buf)
		} else {
			err = errw
		}
	}
	return err
}

func JSONValue(jobj map[string]interface{}, key string) (val string, err error) {
	ival, ok := jobj[key]
	if ok {
		val, ok = ival.(string)
	}
	if !ok {
		err = errors.New("Error retrieving JSON value for: " + key)
	}
	return val, err
}
