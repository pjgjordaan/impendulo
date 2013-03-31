package utils

import (
	"bytes"
	"errors"
	"github.com/disco-volante/intlola/db"
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
const LOG_DIR = "logs"
var BASE_DIR = ".intlola"
var logger *log.Logger

func init() {
	cur, err := user.Current()
	if err == nil {
		temp := cur.HomeDir + SEP + BASE_DIR
		BASE_DIR = ""
		MkDir(temp)
		BASE_DIR = temp
	} else {
		MkDir("")
	}
	now := time.Now()
	y, m, d := now.Date()
	dir := LOG_DIR + SEP + strconv.Itoa(y) + SEP + m.String() + SEP + strconv.Itoa(d)
	MkDir(dir)
	fo, err := os.Create(BASE_DIR + SEP + dir + SEP + time.Now().String() + ".log")
	if err != nil {
		panic(err)
	}
	logger = log.New(fo, "Inlola server log >> ", log.LstdFlags)
}

func WriteFile(file string, data *bytes.Buffer) error {
	return ioutil.WriteFile(BASE_DIR+SEP+file, data.Bytes(), FPERM)
}

func ReadFile(fname string) ([]byte, error) {
	return ioutil.ReadFile(BASE_DIR + SEP + fname)
}

func AddUsers(fname string) error {
	users, err := ReadUsers(fname)
	if err == nil {
		err = db.AddUsers(users...)
	}
	return err
}

func ReadUsers(fname string) (users []*db.UserData, err error) {
	data, err := ioutil.ReadFile(fname)
	if err == nil {
		buff := bytes.NewBuffer(data)
		line, err := buff.ReadString(byte('\n'))
		users = make([]*db.UserData, 100, 1000)
		i := 0
		for err == nil {
			vals := strings.Split(line, ":")
			user := strings.TrimSpace(vals[0])
			pword := strings.TrimSpace(vals[1])
			hash, salt := Hash(pword)
			data := db.NewUser(user, hash, salt)
			if i == len(users) {
				users = append(users, data)
			} else {
				users[i] = data
			}
			i++
			line, err = buff.ReadString(byte('\n'))
		}
		if err == io.EOF {
			err = nil
			if i < len(users) {
				users = users[:i]
			}
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

