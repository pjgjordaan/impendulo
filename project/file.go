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
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/util/errors"
	"labix.org/v2/mgo/bson"

	"reflect"
	"strconv"
	"strings"
)

type (
	Type string
	//File stores a single file's data from a submission.
	File struct {
		Id       bson.ObjectId `bson:"_id"`
		SubId    bson.ObjectId `bson:"subid"`
		Name     string        `bson:"name"`
		Package  string        `bson:"package"`
		Type     Type          `bson:"type"`
		Time     int64         `bson:"time"`
		Data     []byte        `bson:"data"`
		Results  bson.M        `bson:"results"`
		Comments []*Comment    `bson:"comments"`
	}
)

const (
	SRC     Type = "src"
	LAUNCH  Type = "launch"
	ARCHIVE Type = "archive"
	TEST    Type = "test"
)

//String
func (f *File) String() string {
	return "Type: project.File; Id: " + f.Id.Hex() +
		"; SubId: " + f.SubId.Hex() + "; Name: " + f.Name +
		"; Package: " + f.Package + "; Type: " + string(f.Type) +
		"; Time: " + util.Date(f.Time)
}

//Equals
func (f *File) Equals(cf *File) bool {
	if reflect.DeepEqual(f, cf) {
		return true
	}
	return cf != nil && f.String() == cf.String() && bytes.Equal(f.Data, cf.Data)
}

//Same
func (f *File) Same(cf *File) bool {
	return f.Id == cf.Id
}

//CanProcess returns whether a file is meant to be processed.
func (f *File) CanProcess() bool {
	return f.Type == SRC || f.Type == ARCHIVE || f.Type == TEST
}

//NewFile
func NewFile(sid bson.ObjectId, m map[string]interface{}, d []byte) (*File, error) {
	tp, e := convert.GetString(m, TYPE)
	if e != nil && errors.IsCastError(e) {
		return nil, e
	}
	n, e := convert.GetString(m, NAME)
	if e != nil {
		return nil, e
	}
	p, e := convert.GetString(m, PKG)
	if e != nil {
		return nil, e
	}
	t, e := convert.GetInt64(m, TIME)
	if e != nil {
		return nil, e
	}
	return &File{Id: bson.NewObjectId(), SubId: sid, Data: d, Type: Type(tp), Name: n, Package: p, Time: t, Comments: []*Comment{}}, nil
}

//NewArchive
func NewArchive(sid bson.ObjectId, d []byte) *File {
	return &File{Id: bson.NewObjectId(), SubId: sid, Data: d, Type: ARCHIVE, Comments: []*Comment{}}
}

//ParseName retrieves file metadata encoded in a file name.
//These file names must have the format:
//[[<package descriptor>"_"]*<file name>"_"]<time in nanoseconds>
//"_"<file number in current submission>"_"<modification char>
//Where values between '[]' are optional, '*' indicates 0 to many,
//values inside '""' are literals and values inside '<>'
//describe the contents at that position.
func ParseName(n string) (*File, error) {
	es := strings.Split(n, "_")
	if len(es) < 3 {
		return nil, fmt.Errorf("Encoded name %q does not have enough parameters.", n)
	}
	mod := es[len(es)-1]
	ni := 3
	if len(es[len(es)-2]) > 10 {
		ni = 2
	}
	ts := es[len(es)-ni]
	var t int64
	var e error
	if len(ts) == 13 {
		if t, e = strconv.ParseInt(ts, 10, 64); e != nil {
			return nil, fmt.Errorf("%s in %s could not be parsed as an int", ts, n)
		}
	} else if ts[0] == '2' && len(ts) == 17 {
		ct, e := util.CalcTime(ts)
		if e != nil {
			return nil, e
		}
		t = util.GetMilis(ct)
	} else {
		return nil, fmt.Errorf("unknown time format %s in %s.", ts, n)
	}
	var fn, pkg string
	if len(es) > ni {
		ni++
		p := len(es) - ni
		fn = es[p]
		for i := 0; i < p; i++ {
			pkg += es[i]
			if i < p-1 {
				pkg += "."
			}
			if isOutFolder(es[i]) {
				pkg = ""
			}
		}
	}
	var tp Type
	if strings.HasSuffix(fn, JSRC) {
		tp = SRC
	} else if mod == "l" {
		tp = LAUNCH
	} else {
		return nil, fmt.Errorf("unsupported file type in name %s", n)
	}
	return &File{Id: bson.NewObjectId(), Type: tp, Name: fn, Package: pkg, Time: t, Comments: []*Comment{}}, nil
}

//isOutFolder
func isOutFolder(a string) bool {
	return a == SRC_DIR || a == BIN_DIR
}

func (f *File) Rename(n string) {
	if !strings.HasSuffix(n, "java") {
		n += ".java"
	}
	on, _ := util.Extension(f.Name)
	nn, _ := util.Extension(n)
	f.Name = n
	f.Data = bytes.Replace(f.Data, []byte(on), []byte(nn), -1)
}
