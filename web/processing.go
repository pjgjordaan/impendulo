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

package web

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"

	"io/ioutil"

	"labix.org/v2/mgo/bson"

	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

//ReadFormFile reads a file's name and data from a request form.
func ReadFormFile(r *http.Request, n string) (string, []byte, error) {
	f, h, e := r.FormFile(n)
	if e != nil {
		return "", nil, e
	}
	d, e := ioutil.ReadAll(f)
	if e != nil {
		return "", nil, e
	}
	return h.Filename, d, nil
}

//GetStrings retrieves a string value from a request form.
func GetStrings(r *http.Request, n string) ([]string, error) {
	if r.Form == nil {
		if e := r.ParseForm(); e != nil {
			return nil, e
		}
	}
	return r.Form[n], nil
}

//GetString retrieves a string value from a request form.
func GetString(r *http.Request, n string) (string, error) {
	v := r.FormValue(n)
	if strings.TrimSpace(v) != "" {
		return v, nil
	}
	if r.Form == nil {
		if e := r.ParseForm(); e != nil {
			return "", e
		}
	}
	vs, ok := r.Form[n]
	if !ok || len(vs) == 0 || vs[0] == "" {
		return "", fmt.Errorf("invalid value for %s", n)
	}
	return vs[0], nil
}

//getIndex
func getIndex(r *http.Request, n string, maxSize int) (int, error) {
	i, e := convert.Int(r.FormValue(n))
	if e != nil {
		return -1, e
	}
	if i > maxSize {
		return 0, nil
	} else if i < 0 {
		return maxSize, nil
	}
	return i, nil
}

func getId(req *http.Request, ident, name string) (id bson.ObjectId, msg string, err error) {
	id, err = convert.Id(req.FormValue(ident))
	if err != nil {
		msg = fmt.Sprintf("Could not read %s.", name)
	}
	return
}

//getActiveUser
func getActiveUser(ctx *Context) (user, msg string, err error) {
	user, err = ctx.Username()
	if err != nil {
		msg = "Could not retrieve user."
	}
	return
}

func getTestType(r *http.Request) (junit.Type, error) {
	v := strings.ToLower(r.FormValue("testtype"))
	switch v {
	case "default":
		return junit.DEFAULT, nil
	case "admin":
		return junit.ADMIN, nil
	case "user":
		return junit.USER, nil
	default:
		return -1, fmt.Errorf("unsupported test type %s", v)
	}
}

func ServePath(u *url.URL, src string) (string, error) {
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	ext, e := filepath.Rel("/"+filepath.Base(src), path.Clean(u.Path))
	if e != nil {
		return "", e
	}
	sp := filepath.Join(src, ext)
	if util.IsDir(sp) && !strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path + "/"
	}
	return sp, nil
}

func credentials(r *http.Request) (string, string, error) {
	u, e := GetString(r, "user-id")
	if e != nil {
		return "", "", e
	}
	p, e := GetString(r, "password")
	if e != nil {
		return "", "", e
	}
	return u, p, nil
}

//Snapshots retrieves snapshots of a given file in a submission.
func Snapshots(sid bson.ObjectId, n string) ([]*project.File, error) {
	fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: n}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	if len(fs) == 0 {
		return nil, fmt.Errorf("no files found with name %q.", n)
	}
	return fs, nil
}

//getFile
func getFile(id bson.ObjectId) (*project.File, error) {
	return db.File(bson.M{db.ID: id}, bson.M{db.NAME: 1, db.TIME: 1})
}

//projectName retrieves the project name associated with the project identified by id.
func projectName(i interface{}) (string, error) {
	id, e := convert.Id(i)
	if e != nil {
		return "", e
	}
	p, e := db.Project(bson.M{db.ID: id}, bson.M{db.NAME: 1})
	if e != nil {
		return "", e
	}
	return p.Name, nil
}

func projectUsernames(pid bson.ObjectId) ([]string, error) {
	ss, e := db.Submissions(bson.M{db.PROJECTID: pid}, nil)
	if e != nil {
		return nil, e
	}
	ns := make([]string, 0, len(ss))
	a := make(map[string]bool)
	for _, s := range ss {
		if a[s.User] {
			continue
		}
		ns = append(ns, s.User)
		a[s.User] = true
	}
	return ns, nil
}
