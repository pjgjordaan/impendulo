package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	myuser "github.com/godfried/cabanga/user"
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
	"labix.org/v2/mgo/bson"
	"encoding/gob"
	"errors"
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
func ReadUsers(fname string) ([]*myuser.User, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(data)
	line, err := buff.ReadString(byte('\n'))
	users := make([]*myuser.User, 100, 1000)
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
func ReadJSON(r io.Reader) (map[string]interface{}, error) {
	read, err := ReadData(r)
	if err != nil {
		return nil, err
	} 
	var holder interface{}
	err = json.Unmarshal(read, &holder)
	if err != nil {
		return nil, err
	}
	jmap, ok := holder.(map[string]interface{})
	if !ok {
		return nil, errors.New("Failed to cast to JSON: "+string(read))
	}
	return jmap, nil
}

/*
Read data from reader until io.EOF or term is encountered. 
*/
func ReadData(r io.Reader)([]byte, error) {
	buffer := new(bytes.Buffer)
	eof := []byte("eof")
	p := make([]byte, 2048)
	receiving := true
	for receiving {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if bytes.HasSuffix(read, eof) {
			read = read[:len(read)-len(eof)]
			receiving = false
		} 
		if err != nil {
			return nil, err
		}
		buffer.Write(read)
	}
	return buffer.Bytes(), nil
}

/*
Saves a file in a given directory. Creates directory if it doesn't exist.
*/
func SaveFile(dir, fname string, data []byte) error {
	err := os.MkdirAll(dir, DPERM)
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
func UnZip(data []byte) (map[string][]byte, error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		return nil, err
	}
	extracted := make(map[string][]byte)
	for _, zf := range zr.File {
		frc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		if !zf.FileInfo().IsDir() {
			extracted[zf.FileInfo().Name()] = ReadBytes(frc)
		}
		frc.Close()		
	}
	return extracted, nil
}


func Zip(files map[string] []byte)([]byte, error){
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, err
		}
	}
	err := w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ReadBytes(r io.Reader) []byte {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return make([]byte, 0)
	}
	return buffer.Bytes()
}

func LoadMap(fname string) (map[bson.ObjectId]bool, error){
	f, err := os.Open(filepath.Join(BASE_DIR, fname))
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(f)
	var mp map[bson.ObjectId]bool
	err = dec.Decode(&mp)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

/*
Saves map to the filesystem.
*/
func SaveMap(mp map[bson.ObjectId]bool, fname string) error {
	f, err := os.Create(filepath.Join(BASE_DIR, fname))
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	return enc.Encode(&mp)
}



func Merge(m1 map[bson.ObjectId]bool, m2 map[bson.ObjectId]bool){
	for k, v := range m2{
		m1[k] = v
	}
} 
