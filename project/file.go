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
	"github.com/godfried/impendulo/util/milliseconds"
	"labix.org/v2/mgo/bson"

	"reflect"
	"strconv"
	"strings"
)

type (
	Type string
	//File stores a single file's data from a submission.
	File struct {
		Id        bson.ObjectId `bson:"_id"`
		SubId     bson.ObjectId `bson:"subid"`
		Name      string        `bson:"name"`
		Package   string        `bson:"package"`
		Type      Type          `bson:"type"`
		Time      int64         `bson:"time"`
		Data      []byte        `bson:"data"`
		Results   bson.M        `bson:"results"`
		Comments  []*Comment    `bson:"comments"`
		TestCases int           `bson:"testcases"`
	}
	Files []*File
)

const (
	SRC     Type = "src"
	LAUNCH  Type = "launch"
	ARCHIVE Type = "archive"
	TEST    Type = "test"
	ALL     Type = "all"
)

func (fs Files) Less(i, j int) bool {
	return fs[i].Time >= fs[j].Time
}

func (fs Files) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}

func (fs Files) Len() int {
	return len(fs)
}

func ParseType(n string) (Type, error) {
	n = strings.ToLower(n)
	switch n {
	case "source":
		return SRC, nil
	case "src", "launch", "archive", "test", "all":
		return Type(n), nil
	default:
		return Type(""), fmt.Errorf("unknown file type %s", n)
	}
}

func (t Type) Title() string {
	if t == SRC {
		return "Source"
	}
	return util.Title(string(t))
}

func (t Type) String() string {
	if t == SRC {
		return "source"
	}
	return string(t)
}

//String
func (f *File) String() string {
	return "Type: project.File; Id: " + f.Id.Hex() +
		"; SubId: " + f.SubId.Hex() + "; Name: " + f.Name +
		"; Package: " + f.Package + "; Type: " + string(f.Type) +
		"; Time: " + milliseconds.DateTimeString(f.Time)
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
	if e != nil && errors.IsCast(e) {
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
	t, e := parseTime(ts)
	if e != nil {
		return nil, e
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

func parseTime(ts string) (int64, error) {
	if len(ts) == 13 {
		return strconv.ParseInt(ts, 10, 64)
	} else if ts[0] == '2' && len(ts) == 17 {
		return milliseconds.Parse(ts)
	}
	return 0, fmt.Errorf("unknown time format %s.", ts)
}

func ParseFile(n string, d []byte) (*File, error) {
	f, e := ParseName(n)
	if e != nil {
		return nil, e
	}
	f.Data = d
	if pkg := util.GetPackageB(f.Data); pkg != f.Package {
		f.ChangePackage(pkg, f.Package)
	}
	return f, nil
}

//isOutFolder
func isOutFolder(a string) bool {
	return a == SRC_DIR || a == BIN_DIR
}

func (f *File) Rename(p, n string) {
	if n != f.Name {
		if !strings.HasSuffix(n, "java") {
			n += ".java"
		}
		on, _ := util.Extension(f.Name)
		nn, _ := util.Extension(n)
		f.Data = bytes.Replace(f.Data, []byte(on), []byte(nn), -1)
		f.Name = n
	}
	if p != f.Package {
		f.ChangePackage(f.Package, p)
	}
}

func (f *File) ChangePackage(o, n string) {
	f.Data = bytes.Replace(f.Data, []byte("package "+o+";"), []byte("package "+n+";"), -1)
	f.Package = n
}

func (f *File) LoadComments() []*Comment {
	return f.Comments
}

func (f *File) HasCharts() bool {
	return f.Type == SRC
}
