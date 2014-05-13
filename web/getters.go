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
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	//Args represents arguments passed to html templates or to template.Execute.
	Args   map[string]interface{}
	Getter func(r *http.Request, c *Context) (Args, string, error)
)

var (
	getters map[string]Getter
)

//Getters retrieves all getters
func Getters() map[string]Getter {
	if getters == nil {
		getters = defaultGetters()
	}
	return getters
}

//defaultGetters loads the default getters.
func defaultGetters() map[string]Getter {
	return map[string]Getter{
		"configview":    configView,
		"displayresult": displayResult, "getfiles": getFiles,
		"submissionschartview": submissionsChartView, "getsubmissions": getSubmissions,
	}
}

//GenerateGets loads post request handlers and adds them to the router.
func GenerateGets(r *pat.Router, gets map[string]Getter, views map[string]string) {
	for n, f := range gets {
		h := f.CreateGet(n, views[n])
		p := "/" + n
		r.Add("GET", p, Handler(h)).Name(n)
	}
}

func (g Getter) CreateGet(name, view string) Handler {
	return func(w http.ResponseWriter, r *http.Request, c *Context) error {
		a, m, e := g(r, c)
		if m != "" {
			c.AddMessage(m, e != nil)
		}
		if e != nil {
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return e
		}
		t, e := convert.GetStrings(a, "templates")
		if e != nil {
			c.AddMessage("Could not load page.", true)
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return e
		}
		delete(a, "templates")
		c.Browse.View = view
		if c.Browse.View == "home" {
			c.Browse.SetLevel(name)
		}
		a["ctx"] = c
		return T(append(t, getNav(c))...).Execute(w, a)
	}
}

//configView loads a tool's configuration page.
func configView(r *http.Request, c *Context) (Args, string, error) {
	t, e := GetString(r, "tool")
	if e != nil {
		t = "none"
	}
	return Args{"tool": t, "templates": []string{"configview", toolTemplate(t)}},
		"", nil
}

//getSubmissions displays a list of submissions.
func getSubmissions(r *http.Request, c *Context) (Args, string, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	var m bson.M
	if !c.Browse.IsUser {
		m = bson.M{db.PROJECTID: c.Browse.Pid}
	} else {
		m = bson.M{db.USER: c.Browse.Uid}
	}
	s, e := db.Submissions(m, nil, "-"+db.TIME)
	if e != nil {
		return nil, "Could not load submissions.", e
	}
	t := make([]string, 1)
	if c.Browse.IsUser {
		t[0] = "usersubmissionresult"
	} else {
		t[0] = "projectsubmissionresult"
	}
	return Args{"subRes": s, "templates": t}, "", nil
}

//getFiles diplays information about files.
func getFiles(r *http.Request, c *Context) (Args, string, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	f, e := _fileinfos(c.Browse.Sid)
	if e != nil {
		return nil, "Could not retrieve files.", e
	}
	if len(f) == 1 {
		c.Browse.File = f[0].Name
		return displayResult(r, c)
	}
	return Args{"fileInfo": f, "templates": []string{"fileresult"}}, "", nil
}

//displayResult displays a tool's result.
func displayResult(r *http.Request, c *Context) (Args, string, error) {
	a, e := _displayResult(r, c)
	if e != nil {
		return nil, "Could not load results.", e
	}
	return a, "", nil
}

func _displayResult(r *http.Request, c *Context) (Args, error) {
	e := c.Browse.Update(r)
	if e != nil {
		return nil, e
	}
	fs, e := Snapshots(c.Browse.Sid, c.Browse.File)
	if e != nil {
		return nil, e
	}
	cf, e := getFile(fs[c.Browse.Current].Id)
	if e != nil {
		return nil, e
	}
	rs, e := db.ResultNames(c.Browse.Sid, c.Browse.File)
	if e != nil {
		return nil, e
	}
	cr, e := GetResult(c.Browse.Result, cf.Id)
	if e != nil {
		return nil, e
	}
	nf, e := getFile(fs[c.Browse.Next].Id)
	if e != nil {
		return nil, e
	}
	nr, e := GetResult(c.Browse.Result, nf.Id)
	if e != nil {
		return nil, e
	}
	t := []string{"analysisview", "pager", ""}
	if !isError(cr) || isError(nr) {
		t[2] = cr.Template()
	} else {
		t[2] = nr.Template()
	}
	return Args{
		"files": fs, "currentFile": cf, "currentResult": cr, "results": rs,
		"nextFile": nf, "nextResult": nr, "templates": t,
	}, nil
}

func submissionsChartView(r *http.Request, c *Context) (Args, string, error) {
	if e := c.Browse.Update(r); e != nil {
		return nil, "could not update state", e
	}
	t := make([]string, 1)
	if c.Browse.IsUser {
		t[0] = "usersubmissionchart"
	} else {
		t[0] = "projectsubmissionchart"
	}
	return Args{"templates": t}, "", nil
}
