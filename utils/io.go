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
	"encoding/json"
)

const SEP = string(os.PathSeparator)
const DPERM = 0777
const FPERM = 0666
const DEBUG = true
const LOG_DIR = ".intlola"+SEP+"logs"
var logger *log.Logger

func init() {
	cur, err := user.Current()
	if err == nil {
		y, m, d := time.Now().Date()
		dir := cur.HomeDir + SEP + LOG_DIR + SEP + strconv.Itoa(y) + SEP + m.String() + SEP + strconv.Itoa(d)
		err = os.MkdirAll(dir, DPERM)
		if err == nil{
			fo, err := os.Create(dir + SEP + time.Now().String() + ".log")
			if err == nil{
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

func ReadFile(r io.Reader, term []byte)(buffer *bytes.Buffer, err error){
	buffer = new(bytes.Buffer)
	p := make([]byte, 2048)
	receiving := true
	for receiving {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if bytes.HasSuffix(read, term) || err != nil{
			read = read[:len(read)-len(term)]
			receiving = false
		} 
		if err == nil || err == io.EOF{
			buffer.Write(read)
		}
	}
	if err == io.EOF{
		err = nil
	}
	return buffer, err
}