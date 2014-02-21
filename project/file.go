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

package project

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	Type string
	//File stores a single file's data from a submission.
	File struct {
		Id      bson.ObjectId "_id"
		SubId   bson.ObjectId "subid"
		Name    string        "name"
		Package string        "package"
		Type    Type          "type"
		Time    int64         "time"
		Data    []byte        "data"
		Results bson.M        "results"
	}
)

const (
	SRC     Type = "src"
	LAUNCH  Type = "launch"
	ARCHIVE Type = "archive"
	TEST    Type = "test"
)

//String
func (this *File) String() string {
	return "Type: project.File; Id: " + this.Id.Hex() +
		"; SubId: " + this.SubId.Hex() + "; Name: " + this.Name +
		"; Package: " + this.Package + "; Type: " + string(this.Type) +
		"; Time: " + util.Date(this.Time)
}

//Equals
func (this *File) Equals(that *File) bool {
	if reflect.DeepEqual(this, that) {
		return true
	}
	return that != nil &&
		this.String() == that.String() &&
		bytes.Equal(this.Data, that.Data)
}

//Same
func (this *File) Same(that *File) bool {
	return this.Id == that.Id
}

//CanProcess returns whether a file is meant to be processed.
func (this *File) CanProcess() bool {
	return this.Type == SRC || this.Type == ARCHIVE
}

//NewFile
func NewFile(subId bson.ObjectId, info map[string]interface{}, data []byte) (file *File, err error) {
	id := bson.NewObjectId()
	file = &File{Id: id, SubId: subId, Data: data}
	tipe, err := util.GetString(info, TYPE)
	if err != nil && util.IsCastError(err) {
		return
	}
	file.Type = Type(tipe)
	file.Name, err = util.GetString(info, NAME)
	if err != nil {
		return
	}
	file.Package, err = util.GetString(info, PKG)
	if err != nil {
		return
	}
	file.Time, err = util.GetInt64(info, TIME)
	return
}

//NewArchive
func NewArchive(subId bson.ObjectId, data []byte) *File {
	id := bson.NewObjectId()
	return &File{
		Id:    id,
		SubId: subId,
		Data:  data,
		Type:  ARCHIVE,
	}
}

//ParseName retrieves file metadata encoded in a file name.
//These file names must have the format:
//[[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>
//"_"<file number in current submission>"_"<modification char>
//Where values between '[]' are optional, '*' indicates 0 to many,
//values inside '""' are literals and values inside '<>'
//describe the contents at that position.
func ParseName(name string) (file *File, err error) {
	elems := strings.Split(name, "_")
	if len(elems) < 3 {
		err = fmt.Errorf("Encoded name %q does not have enough parameters.", name)
		return
	}
	file = new(File)
	file.Id = bson.NewObjectId()
	mod := elems[len(elems)-1]
	nextIndex := 3
	if len(elems[len(elems)-2]) > 10 {
		nextIndex = 2
	}
	timeString := elems[len(elems)-nextIndex]
	if len(timeString) == 13 {
		file.Time, err = strconv.ParseInt(timeString, 10, 64)
		if err != nil {
			err = fmt.Errorf(
				"%s in name %s could not be parsed as an int.",
				timeString, name)
			return
		}
	} else if timeString[0] == '2' && len(timeString) == 17 {
		var t time.Time
		t, err = util.CalcTime(timeString)
		if err != nil {
			return
		}
		file.Time = util.GetMilis(t)
	} else {
		err = fmt.Errorf(
			"Unknown time format %s in %s.",
			timeString, name)
		return
	}
	if len(elems) > nextIndex {
		nextIndex++
		pos := len(elems) - nextIndex
		file.Name = elems[pos]
		for i := 0; i < pos; i++ {
			file.Package += elems[i]
			if i < pos-1 {
				file.Package += "."
			}
			if isOutFolder(elems[i]) {
				file.Package = ""
			}
		}
	}
	if strings.HasSuffix(file.Name, JSRC) {
		file.Type = SRC
	} else if mod == "l" {
		file.Type = LAUNCH
	} else {
		err = fmt.Errorf("Unsupported file type in name %s", name)
	}
	return
}

//isOutFolder
func isOutFolder(arg string) bool {
	return arg == SRC_DIR || arg == BIN_DIR
}
