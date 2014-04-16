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

//GetInt retrieves an integer value from a request form.
func GetInt(r *http.Request, n string) (int, error) {
	return util.Int(r.FormValue(n))
}

//GetInt64 retrieves an integer value from a request form.
func GetInt64(r *http.Request, n string) (int64, error) {
	return util.Int64(r.FormValue(n))
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
	if strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("invalid value for %s", n)
	}
	return v, nil
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
func getProjectId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "projectid", "project")
}

//getSubId
func getSubId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "subid", "submission")
}

//getSkeletonId
func getSkeletonId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "skeletonid", "skeleton")
}

//getFileId
func getFileId(req *http.Request) (bson.ObjectId, string, error) {
	return getId(req, "fileid", "file")
}

func getId(req *http.Request, ident, name string) (id bson.ObjectId, msg string, err error) {
	id, err = util.ReadId(req.FormValue(ident))
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

//getUserId
func getUserId(req *http.Request) (userid, msg string, err error) {
	userid, err = GetString(req, "userid")
	if err != nil {
		msg = "Could not read user."
	}
	return
}

//getString.
func getString(r *http.Request, n string) (string, string, error) {
	s, e := GetString(r, n)
	if e != nil {
		return "", fmt.Sprintf("could not read %s", n), e
	}
	return s, "", e
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

//getCredentials
func getCredentials(r *http.Request) (u, p, m string, e error) {
	if u, m, e = getString(r, "username"); e != nil {
		return
	}
	p, m, e = getString(r, "password")
	return
}

//Snapshots retrieves snapshots of a given file in a submission.
func Snapshots(sid bson.ObjectId, n string, t project.Type) ([]*project.File, error) {
	fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: n, db.TYPE: t}, bson.M{db.DATA: 0}, db.TIME)
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
	id, e := util.ReadId(i)
	if e != nil {
		return "", e
	}
	p, e := db.Project(bson.M{db.ID: id}, bson.M{db.NAME: 1})
	if e != nil {
		return "", e
	}
	return p.Name, nil
}

func childFile(sid bson.ObjectId, n string) (*project.File, error) {
	pfs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: n}, bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	if len(pfs) == 0 {
		return nil, fmt.Errorf("no files named %s in submission %s", n, sid)
	}
	switch pfs[0].Type {
	case project.SRC:
		s, e := db.Submission(bson.M{db.ID: sid}, nil)
		if e != nil {
			return nil, e
		}
		ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId, db.TYPE: junit.USER}, bson.M{db.NAME: 1})
		if e != nil {
			return nil, e
		}
		for _, t := range ts {
			tn, _ := util.Extension(t.Name)
			for _, pf := range pfs {
				f, e := db.File(bson.M{db.SUBID: sid, db.TYPE: project.TEST, db.NAME: t.Name, db.RESULTS + "." + tn + "-" + pf.Id.Hex(): bson.M{db.EXISTS: true}}, bson.M{db.DATA: 0})
				if e != nil {
					continue
				}
				return f, nil
			}
		}
		return nil, fmt.Errorf("no tests found for file %s in submission %s", n, sid)
	case project.TEST:
		for _, t := range pfs {
			for k, _ := range t.Results {
				s := strings.Split(k, "-")
				if len(s) < 2 {
					continue
				}
				id, e := util.ReadId(s[len(s)-1])
				if e != nil {
					continue
				}
				f, e := db.File(bson.M{db.ID: id}, bson.M{db.DATA: 0})
				if e != nil {
					continue
				}
				return f, nil
			}
		}
		return nil, fmt.Errorf("No file name found for submission %s", sid.Hex())
	default:
		return nil, fmt.Errorf("unsupported type %s", pfs[0].Type)
	}
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
