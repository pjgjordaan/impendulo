package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/gob"
	"errors"
	"github.com/disco-volante/intlola/db"
	myuser "github.com/disco-volante/intlola/user"
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
const BASE_DIR = ".intlola"
const LOG_DIR = "logs"

var logger *log.Logger
var logM *sync.Mutex

func init() {
	logM = new(sync.Mutex)
	cur, err := user.Current()
	if err == nil {
		y, m, d := time.Now().Date()
		dir := filepath.Join(cur.HomeDir, BASE_DIR, LOG_DIR, strconv.Itoa(y), m.String(), strconv.Itoa(d))
		err = os.MkdirAll(dir, DPERM)
		if err == nil {
			fo, err := os.Create(filepath.Join(dir, time.Now().String()+".log"))
			if err == nil {
				logger = log.New(fo, "Inlola server log >> ", log.LstdFlags)

			}
		}
	}
	if err != nil {
		panic(err)
	}
}

func AddUsers(fname string) error {
	users, err := ReadUsers(fname)
	if err == nil {
		err = db.AddMany(db.USERS, users...)
	}
	return err
}

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

func ReadFile(r io.Reader, term []byte) (buffer *bytes.Buffer, err error) {
	buffer = new(bytes.Buffer)
	p := make([]byte, 2048)
	receiving := true
	for receiving {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if bytes.HasSuffix(read, term)  {
			read = read[:len(read)-len(term)]
			receiving = false
		} else if err != nil{
			receiving = false
		}
		if err == nil || err == io.EOF {
			buffer.Write(read)
			err = nil
		}
	}
	return buffer, err
}

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

func SaveStruct(dir, fname string, strct interface{}) (err error){
	err = os.MkdirAll(dir, DPERM)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, fname))
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f) 
	err  = enc.Encode(strct)
	return err
}

func ReadStruct(dir, fname string, strct interface{}) (err error){
	f, err := os.Open(filepath.Join(dir, fname))
	if err != nil{
		return err
	}
	dec := gob.NewDecoder(f) 
	err = dec.Decode(strct)
	return err

}

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

func ReadZip(data []byte) (extracted map[string][]byte, err error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	extracted = make(map[string][]byte)
	if err == nil {
		for _, zf := range zr.File {
			frc, err := zf.Open()
			if err == nil {
				if !zf.FileInfo().IsDir() {
					extracted[zf.FileInfo().Name()] = getBytes(frc)
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

func getBytes(r io.Reader) []byte {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
