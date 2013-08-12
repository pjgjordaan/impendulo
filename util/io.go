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

type IOError struct {
	origin interface{}
	tipe   string
	err    error
}

func (this *IOError) Error() string {
	return fmt.Sprintf(`Encountered error %q while %s %q.`,
		this.err, this.tipe, this.origin)
}

//BaseDir retrieves the Impendulo directory.
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
			return nil, &IOError{r, "reading from", err}
		} else if bytes.HasSuffix(read, eof) {
			read = read[:len(read)-len(eof)]
			busy = false
		}
		buffer.Write(read)
	}
	return buffer.Bytes(), nil
}

//SaveFile saves a file (given as a []byte) as fname.
func SaveFile(fname string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(fname), DPERM)
	if err != nil {
		return &IOError{fname, "creating", err}
	}
	f, err := os.Create(fname)
	if err != nil {
		return &IOError{fname, "creating", err}
	}
	_, err = f.Write(data)
	if err != nil {
		return &IOError{fname, "writing to", err}
	}
	return nil
}

//ReadBytes reads bytes from a reader until io.EOF is encountered.
//If the reader can't be read an empty []byte is returned.
func ReadBytes(r io.Reader) (data []byte) {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r)
	if err != nil {
		data = make([]byte, 0)
	} else {
		data = buffer.Bytes()
	}
	return
}

//GetPackage retrieves the package name from a Java source file.
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

type copier struct {
	dest, src string
}

func (this *copier) copyFile(path string, f os.FileInfo, err error) error {
	if err != nil || f == nil {
		return err
	}
	return this.copy(path, f)
}

func (this *copier) copy(path string, f os.FileInfo) (err error) {
	destPath, err := filepath.Rel(this.src, path)
	if err != nil {
		return
	}
	destPath = filepath.Join(this.dest, destPath)
	if f.IsDir() {
		err = os.MkdirAll(destPath, os.ModePerm)
		return
	}
	srcFile, err := os.Open(path)
	if err != nil {
		return
	}
	destFile, err := os.Create(destPath)
	if err != nil {
		return
	}
	_, err = io.Copy(destFile, srcFile)
	return
}

//Copy copies the contents of src to dest.
func Copy(dest, src string) error {
	c := &copier{dest, src}
	return filepath.Walk(src, c.copyFile)
}

//Exists checks whether a given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
