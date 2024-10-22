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
	"code.google.com/p/gorilla/pat"

	"os"

	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/mongo"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"

	"io/ioutil"

	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	Downloader func(*http.Request) (path string, err error)
)

var (
	downloaders map[string]Downloader
)

func Downloaders() map[string]Downloader {
	if downloaders == nil {
		downloaders = map[string]Downloader{
			"skeleton.zip": LoadSkeleton,
			"intlola.zip":  LoadIntlola,
			"exportdb.zip": ExportData,
			"test.zip":     LoadTest,
		}
	}
	return downloaders
}

//GenerateDownloads
func GenerateDownloads(r *pat.Router, d map[string]Downloader) {
	for n, f := range d {
		r.Add("GET", "/"+n, Handler(f.CreateDownload())).Name(n)
	}
}

//CreateDownload.
func (d Downloader) CreateDownload() Handler {
	return func(w http.ResponseWriter, r *http.Request, c *context.C) error {
		p, e := d(r)
		if e != nil {
			c.AddMessage("could not load file for downloading.", true)
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return e
		}
		http.ServeFile(w, r, p)
		return os.Remove(p)
	}
}

//LoadSkeleton makes a project skeleton available for download.
func LoadTest(r *http.Request) (string, error) {
	id, e := convert.Id(r.FormValue("test-id"))
	if e != nil {
		return "", e
	}
	t, e := db.JUnitTest(bson.M{db.ID: id}, nil)
	if e != nil {
		return "", e
	}
	z, e := util.ZipMap(map[string][]byte{t.Name: t.Test})
	if e != nil {
		return "", e
	}
	//Save file to filesystem and return path to it.
	return util.SaveTemp(z)
}

//LoadSkeleton makes a project skeleton available for download.
func LoadSkeleton(r *http.Request) (string, error) {
	id, e := convert.Id(r.FormValue("skeleton-id"))
	if e != nil {
		return "", e
	}
	s, e := db.Skeleton(bson.M{db.ID: id}, nil)
	if e != nil {
		return "", e
	}
	//Save file to filesystem and return path to it.
	return util.SaveTemp(s.Data)
}

//ExportData
func ExportData(r *http.Request) (string, error) {
	n, e := webutil.String(r, "db")
	if e != nil {
		return "", e
	}
	c, e := webutil.Strings(r, "collections")
	if e != nil {
		return "", e
	}
	p, e := mongo.ExportData(n, c)
	if e != nil {
		return "", e
	}
	return p, nil
}

func LoadIntlola(r *http.Request) (string, error) {
	p, e := config.INTLOLA.Path()
	if e != nil {
		return "", e
	}
	b, e := ioutil.ReadFile(p)
	if e != nil {
		return "", e
	}
	return util.SaveTemp(b)
}
