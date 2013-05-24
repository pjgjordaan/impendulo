package util

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	myuser "github.com/godfried/cabanga/user"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
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

var BASE_DIR = ".intlola"
var LOG_DIR = "logs"
var errLogger, infoLogger *SyncLogger

//init sets up the loggers.
func init() {
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
	errLogger, err = NewLogger(filepath.Join(dir, "E_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
	infoLogger, err = NewLogger(filepath.Join(dir, "I_"+time.Now().String()+".log"))
	if err != nil {
		panic(err)
	}
}

//SyncLogger allows for concurrent logging.
type SyncLogger struct {
	logger *log.Logger
	lock   *sync.Mutex
}

//Log safely logs data to this logger's log file.
func (this SyncLogger) Log(vals ...interface{}) {
	this.lock.Lock()
	this.logger.Print(vals...)
	this.lock.Unlock()
}

//NewLogger creates a new SyncLogger which writes its logs to the specified file.
func NewLogger(fname string) (*SyncLogger, error) {
	fo, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &SyncLogger{log.New(fo, "", log.LstdFlags), new(sync.Mutex)}, nil
}

//Log sends data to be logged to the appropriate logger.
func Log(v ...interface{}) {
	if len(v) > 0 {
		if _, ok := v[0].(error); ok {
			errLogger.Log(v)
		} else {
			infoLogger.Log(v)
		}
	}
}

//ReadUsers reads user configurations from a file.
//It also sets up their passwords.
func ReadUsers(fname string) ([]*myuser.User, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when attempting to read file %q", err, fname)
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

//ReadJSON reads all JSON data from a reader. 
func ReadJSON(r io.Reader) (map[string]interface{}, error) {
	read, err := ReadData(r)
	if err != nil {
		return nil, err
	}
	var holder interface{}
	err = json.Unmarshal(read, &holder)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when unmarshaling data %q", err, read)
	}
	jmap, ok := holder.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Encountered error when attempting to cast %q to json map", holder)
	}
	return jmap, nil
}

//ReadData reads data from a reader until io.EOF or []byte("eof") is encountered. 
func ReadData(r io.Reader) ([]byte, error) {
	buffer := new(bytes.Buffer)
	eof := []byte("eof")
	p := make([]byte, 2048)
	busy := true
	for busy {
		bytesRead, err := r.Read(p)
		read := p[:bytesRead]
		if err == io.EOF {
			busy = false
		} else if err != nil {
			return nil, fmt.Errorf("Encountered error %q while reading from %q", err, r)
		} else if bytes.HasSuffix(read, eof) {
			read = read[:len(read)-len(eof)]
			busy = false
		}
		buffer.Write(read)
	}
	return buffer.Bytes(), nil
}

//SaveFile saves a file (given as a []byte)  in dir.
func SaveFile(dir, fname string, data []byte) error {
	err := os.MkdirAll(dir, DPERM)
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating directory %q", err, dir)
	}
	f, err := os.Create(filepath.Join(dir, fname))
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating file %q", err, fname)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("Encountered error %q while writing data to %q", err, f)
	}
	return nil
}

//Unzip extracts a file (given as a []byte) to dir.
func Unzip(dir string, data []byte) error {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
	}
	for _, zf := range zr.File {
		err = ExtractFile(zf, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

//ExtractFile extracts the contents of a zip.File and saves it in dir.
func ExtractFile(zf *zip.File, dir string) error {
	frc, err := zf.Open()
	if err != nil {
		return fmt.Errorf("Encountered error %q while opening zip file %q", err, zf)
	}
	defer frc.Close()
	path := filepath.Join(dir, zf.Name)
	if zf.FileInfo().IsDir() {
		err = os.MkdirAll(path, DPERM)
		if err != nil {
			return fmt.Errorf("Encountered error %q while creating directory %q", err, path)
		}
	} else {
		err = os.MkdirAll(filepath.Dir(path), DPERM)
		if err != nil {
			return fmt.Errorf("Encountered error %q while creating directory %q", err, path)
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zf.Mode())
		if err != nil {
			return fmt.Errorf("Encountered error %q while opening file %q", err, path)
		}
		defer f.Close()
		_, err = io.Copy(f, frc)
		if err != nil {
			return fmt.Errorf("Encountered error %q while copying from %q to %q", err, frc, f)
		}
	}
	return nil
}

//UnzipToMap reads the contents of a zip file into a map.
//Each file's path is a map key and its data is the associated value. 
func UnzipToMap(data []byte) (map[string][]byte, error) {
	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, int64(br.Len()))
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while creating zip reader from from %q", err, data)
	}
	extracted := make(map[string][]byte)
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		data, err := ExtractBytes(zf)
		if err != nil {
			return nil, err
		}
		extracted[zf.FileInfo().Name()] = data
	}
	return extracted, nil
}

//ExtractBytes extracts data from a zip.File. 
func ExtractBytes(zf *zip.File) ([]byte, error) {
	frc, err := zf.Open()
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while opening zip file %q", err, zf)
	}
	defer frc.Close()
	return ReadBytes(frc), nil
}

//Zip creates a zip archive from a map which has file names as its keys and file contents as its values.
func Zip(files map[string][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			return nil, fmt.Errorf("Encountered error %q while creating file %q in zip %q", err, name, w)
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, fmt.Errorf("Encountered error %q while writing to file %q in zip %q", err, f, w)
		}
	}
	err := w.Close()
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while closing zip %q", err, w)
	}
	return buf.Bytes(), nil
}

//ReadBytes reads bytes from a reader until io.EOF is encountered.
//If the reader can't be read an empty []byte is returned.
func ReadBytes(r io.Reader) []byte {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return make([]byte, 0)
	}
	return buffer.Bytes()
}

//LoadMap loads a map stored in a file.
func LoadMap(fname string) (map[bson.ObjectId]bool, error) {
	f, err := os.Open(filepath.Join(BASE_DIR, fname))
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while opening file %q", err, fname)
	}
	dec := gob.NewDecoder(f)
	var mp map[bson.ObjectId]bool
	err = dec.Decode(&mp)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q while decoding map stored in %q", err, f)
	}
	return mp, nil
}

//SaveMap saves a map to the filesystem.
func SaveMap(mp map[bson.ObjectId]bool, fname string) error {
	f, err := os.Create(filepath.Join(BASE_DIR, fname))
	if err != nil {
		return fmt.Errorf("Encountered error %q while creating file %q", err, fname)
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(&mp)
	if err != nil {
		return fmt.Errorf("Encountered error %q while encoding map %q to file %q", err, mp, fname)
	}
	return nil
}

//Merge adds the contents of m2 to m1. Values in m1 may be overwritten.
func Merge(m1 map[bson.ObjectId]bool, m2 map[bson.ObjectId]bool) {
	for k, v := range m2 {
		m1[k] = v
	}
}
