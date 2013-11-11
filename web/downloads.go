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
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/mongo"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
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
			"exportdb.zip": ExportData,
		}
	}
	return downloaders
}

//GenerateDownloads
func GenerateDownloads(router *pat.Router, downloads map[string]Downloader) {
	for name, fn := range downloads {
		handleFunc := fn.CreateDownload()
		pattern := "/" + name
		router.Add("GET", pattern, Handler(handleFunc)).Name(name)
	}
}

//CreateDownload.
func (this Downloader) CreateDownload() Handler {
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error {
		path, err := this(req)
		if err != nil {
			ctx.AddMessage("Could not load file for downloading.", true)
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		} else {
			http.ServeFile(w, req, path)
		}
		return err
	}
}

//LoadSkeleton makes a project skeleton available for download.
func LoadSkeleton(req *http.Request) (path string, err error) {
	projectId, _, err := getProjectId(req)
	if err != nil {
		return
	}
	name := bson.NewObjectId().Hex()
	base, err := util.BaseDir()
	if err != nil {
		return
	}
	path = filepath.Join(base, "skeletons", name+".zip")
	//If the skeleton is saved for downloading we don't need to store it again.
	if util.Exists(path) {
		return
	}
	p, err := db.Project(bson.M{db.ID: projectId}, nil)
	if err != nil {
		return
	}
	//Save file to filesystem and return path to it.
	err = util.SaveFile(path, p.Skeleton)
	return
}

//ExportData
func ExportData(req *http.Request) (path string, err error) {
	dbName, err := GetString(req, "db")
	if err != nil {
		return
	}
	collections, err := GetStrings(req, "collections")
	if err != nil {
		return
	}
	name := strconv.FormatInt(time.Now().Unix(), 10)
	base, err := util.BaseDir()
	if err != nil {
		return
	}
	path = filepath.Join(base, "exports", name+".zip")
	err = mongo.ExportData(path, dbName, collections)
	return
}
