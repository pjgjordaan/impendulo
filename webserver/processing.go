//Contains processing functions used by handlers.go to add or retrieve data.
package webserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

//ReadFormFile reads a file's name and data from a request form.
func ReadFormFile(req *http.Request, name string) (fname string, data []byte, err error) {
	file, header, err := req.FormFile(name)
	if err != nil {
		return
	}
	fname = header.Filename
	data, err = ioutil.ReadAll(file)
	return
}

//GetInt retrieves an integer value from a request form.
func GetInt(req *http.Request, name string) (found int, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.Atoi(iStr)
	return
}

func GetLines(req *http.Request, name string) []int {
	start, err := GetInt(req, name+"focusstart")
	if err != nil {
		err = nil
		start = 0
	}
	end, err := GetInt(req, name+"focusend")
	if err != nil {
		err = nil
		end = start
	}
	lines := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		lines[i-start] = i
	}
	return lines
}

//GetStrings retrieves a string value from a request form.
func GetStrings(req *http.Request, name string) (vals []string, err error) {
	if req.Form == nil {
		err = req.ParseForm()
		if err != nil {
			return
		}
	}
	vals = req.Form[name]
	return
}

//GetString retrieves a string value from a request form.
func GetString(req *http.Request, name string) (val string, err error) {
	val = req.FormValue(name)
	if strings.TrimSpace(val) == "" {
		err = fmt.Errorf("Invalid value for %s.", name)
	}
	return
}

func getIndex(req *http.Request, name string, maxSize int) (ret int, err error) {
	ret, err = GetInt(req, name)
	if err != nil {
		return
	}
	if ret > maxSize {
		ret = 0
	} else if ret < 0 {
		ret = maxSize
	}
	return
}

func getSelected(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "currentIndex", maxSize)
}

func getNeighbour(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "nextIndex", maxSize)
}
