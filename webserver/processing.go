//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/mongo"
	"github.com/godfried/impendulo/util"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//CloneData
func CloneData(req *http.Request, ctx *Context) (msg string, err error) {
	remote, err := GetString(req, "remote")
	if err != nil {
		msg = "Could not read remote data location."
		return
	}
	err = db.CloneData(remote)
	if err != nil {
		msg = fmt.Sprintf("Could not clone data from %s.", remote)
	}
	return
}

//ImportData
func ImportData(req *http.Request, ctx *Context) (msg string, err error) {
	dbName, err := GetString(req, "db")
	if err != nil {
		msg = "Could not read db to import to."
		return
	}
	var data []byte
	_, data, err = ReadFormFile(req, "data")
	if err != nil {
		msg = "Unable to read data file."
		return
	}
	err = mongo.ImportData(dbName, data)
	if err != nil {
		msg = "Unable to import db data."
	} else {
		msg = "Successfully imported db data."
	}
	return
}

//ExportData
func ExportData(req *http.Request, ctx *Context) (msg string, err error) {
	dbName, err := GetString(req, "db")
	if err != nil {
		msg = "Could not read db to export."
		return
	}
	collections, err := GetStrings(req, "collections")
	if err != nil {
		msg = "Could not read collections to export."
		return
	}
	name := strconv.FormatInt(time.Now().Unix(), 10)
	path := filepath.Join(util.BaseDir(), "exports", name+".zip")
	err = mongo.ExportData(dbName, path, collections)
	if err != nil {
		msg = "Unable to export db data."
	} else {
		msg = "Successfully exported db data."
	}
	return
}

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

//GetInt64 retrieves an integer value from a request form.
func GetInt64(req *http.Request, name string) (found int64, err error) {
	iStr := req.FormValue(name)
	found, err = strconv.ParseInt(iStr, 10, 64)
	return
}

//GetLines retrieves an array of size m-n+1 with values
//starting at n and ending at m where n and m are start and end
//values retrieved from req.
func GetLines(req *http.Request) (lines []int, err error) {
	lineStr, err := GetString(req, "lines")
	if err != nil {
		return
	}
	locs := strings.Split(lineStr, "-")
	start, err := strconv.Atoi(locs[0])
	if err != nil {
		return
	}
	if start < 0 {
		err = fmt.Errorf("Invalid code highlight start value %d.", start)
		return
	}
	end := start
	if len(locs) == 2 {
		end, err = strconv.Atoi(locs[1])
		if err != nil {
			return
		} else if end < start {
			err = fmt.Errorf("Code highlight end value %d less than start %d.", end, start)
			return
		}
	}
	lines = make([]int, end-start+1)
	for i := start; i <= end; i++ {
		lines[i-start] = i
	}
	return
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

//getIndex
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

//getCurrent
func getCurrent(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "current", maxSize)
}

//getNext
func getNext(req *http.Request, maxSize int) (int, error) {
	return getIndex(req, "next", maxSize)
}

//getProjectId
func getProjectId(req *http.Request) (id bson.ObjectId, msg string, err error) {
	id, err = util.ReadId(req.FormValue("project"))
	if err != nil {
		msg = "Could not read project."
	}
	return
}

//getUser
func getUser(ctx *Context) (user, msg string, err error) {
	user, err = ctx.Username()
	if err != nil {
		msg = "Could not retrieve user."
	}
	return
}

//getString.
func getString(req *http.Request, name string) (val, msg string, err error) {
	val, err = GetString(req, name)
	if err != nil {
		msg = fmt.Sprintf("Could not read %s.", name)
	}
	return
}
