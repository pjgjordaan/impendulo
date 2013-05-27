package util

import (
	"bytes"
	"fmt"
	myuser "github.com/godfried/cabanga/user"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const DPERM = 0777
const FPERM = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func BaseDir() string{
	cur, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(cur.HomeDir, ".intlola")
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
