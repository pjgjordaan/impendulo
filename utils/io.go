package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	myuser "github.com/godfried/intlola/user"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const DPERM = 0777
const FPERM = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
const DEBUG = true

var BASE_DIR = ".intlola"
var LOG_DIR = "logs"

var logger *log.Logger
var logM *sync.Mutex

/*
Setup logger.
*/
func init() {
	logM = new(sync.Mutex)
	cur, err := user.Current()
	if err != nil {
		panic(err)
	}
	y, m, d := time.Now().Date()
	BASE_DIR = filepath.Join(cur.HomeDir, BASE_DIR)
	LOG_DIR = filepath.Join(BASE_DIR, LOG_DIR)
	dir := filepath.Join(LOG_DIR, strconv.Itoa(y), m.String(), strconv.Itoa(d))
	err = os.MkdirAll(dir, DPERM)
	if err != nil {
		panic(err)
	}
	fo, err := os.Create(filepath.Join(dir, time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
	logger = log.New(fo, "", log.LstdFlags)
}

/*
Reads users configurations from file. Sets up passwords.
*/
func ReadUsers(fname string) (users []interface{}, err error) {
	data, err := ioutil.ReadFile(fname)
	if err == nil {
		buff := bytes.NewBuffer(data)
		line, err := buff.ReadString(byte('\n'))
		users = make([]interface{}, 100, 1000)
		i := 0
		for err == nil {
			vals := strings.Split(line, ":")
			uname := strings.TrimSpace(vals[0])
			pword := strings.TrimSpace(vals[1])
			hash, salt := Hash(pword)
			data := &myuser.User{uname, hash, salt, myuser.ALL_SUB}
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
		logM.Lock()
		logger.Print(v...)
		logM.Unlock()
	}
}

/*
 Reads all JSON data from reader (maximum of 1024 bytes). 
*/
func ReadJSON(r io.Reader) (jobj map[string]interface{}, err error) {
	buffer := make([]byte, 1024)
	bytesRead, err := r.Read(buffer)
	if err == nil {
		buffer = buffer[:bytesRead]
		var holder interface{}
		err = json.Unmarshal(buffer, &holder)
		if err == nil {
			jobj = holder.(map[string]interface{})
		}
	}
	return jobj, err
}

/*
Read data from reader until io.EOF or term is encountered. 
*/
func ReadData(r io.Reader, term []byte) (buffer *bytes.Buffer, err error) {
	buffer = new(bytes.Buffer)
	p := make([]byte, 2048)
	receiving := true
	for receiving {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if bytes.HasSuffix(read, term) {
			read = read[:len(read)-len(term)]
			receiving = false
		} else if err != nil {
			receiving = false
		}
		if err == nil || err == io.EOF {
			buffer.Write(read)
			err = nil
		}
	}
	return buffer, err
}

/*
Saves a file in a given directory. Creates directory if it doesn't exist.
*/
func SaveFile(dir, fname string, data []byte) (err error) {
	err = os.MkdirAll(dir, DPERM)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, fname))
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

/*
 Unzip a file (data) to a given directory.
*/
func Unzip(dir string, data []byte) (err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err == nil {
		for _, zf := range zr.File {
			frc, err := zf.Open()
			if err == nil {
				path := filepath.Join(dir, zf.Name)
				if zf.FileInfo().IsDir() {
					err = os.MkdirAll(path, zf.Mode())
				} else {
					f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
					if err == nil {
						_, err = io.Copy(f, frc)
						f.Close()
					}
				}
				frc.Close()
			}
			if err != nil {
				break
			}
		}
	}
	return err
}

/*
Read contents of zip file into a map with each file's path being a map
key and data being the value. 
*/
func ReadZip(data []byte) (extracted map[string][]byte, err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	extracted = make(map[string][]byte)
	if err == nil {
		for _, zf := range zr.File {
			frc, err := zf.Open()
			if err == nil {
				if !zf.FileInfo().IsDir() {
					extracted[zf.FileInfo().Name()] = ReadBytes(frc)
				}
				frc.Close()
			}
			if err != nil {
				break
			}
		}
	}
	return extracted, err
}

func ReadBytes(r io.Reader) []byte {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return make([]byte, 0)
	}
	return buffer.Bytes()
}
