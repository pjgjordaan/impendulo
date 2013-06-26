package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const DPERM = 0777
const FPERM = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

func BaseDir() string {
	cur, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(cur.HomeDir, ".impendulo")
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

func GetPackage(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "package" {
			scanner.Scan()
			return strings.Split(scanner.Text(), ";")[0]
		}
	}
	return ""
}

func GenHTML(dir, name string, data []byte) (string, error) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while creating directory %q", err, dir)
	}
	name = filepath.Join(dir, name+".html")
	f, err := os.Create(name)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while creating file %q", err, name)
	}
	_, err = f.Write(data)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while writing data to %q", err, f)
	}
	return "/" + name, nil
}
type copier struct{
	dest, src string
}

func (this *copier) copyFile(path string, f os.FileInfo, err error) error{
	destPath, err := filepath.Rel(this.src, path)
	if err != nil{
		return err
	}
	destPath = filepath.Join(this.dest, destPath)
	if f == nil{
		fmt.Println(err, path, f)
		return nil
	} else if f.IsDir(){
		return os.MkdirAll(destPath, os.ModePerm)
	} else{
		srcFile, err := os.Open(path)
		if err != nil{
			return err
		}
		destFile, err := os.Create(destPath)
		if err != nil{
			return err
		}
		_, err = io.Copy(destFile, srcFile)
		return err
	}
}

func Copy(dest, src string) error{
	c := &copier{dest, src}
	return filepath.Walk(src, c.copyFile)
}