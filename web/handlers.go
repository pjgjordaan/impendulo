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
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"net/http"
)

var (
	store sessions.Store
)

const (
	LOG_HANDLERS = "webserver/handlers.go"
)

type (
	//Handler is used to handle incoming requests.
	//It allows for better session management.
	Handler func(http.ResponseWriter, *http.Request, *Context) error
)

func init() {
	auth, enc, err := util.CookieKeys()
	if err != nil {
		panic(err)
	}
	store = sessions.NewCookieStore(auth, enc)
}

//ServeHTTP loads a the current session, handles  the request and
//then stores the session.
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		util.Log(err, LOG_HANDLERS)
	}
	//Load our context from session
	ctx := LoadContext(sess)
	buf := new(HttpBuffer)
	err = CheckAccess(req.URL.Path, ctx, Permissions())
	if err != nil {
		ctx.AddMessage(err.Error(), true)
		http.Redirect(buf, req, getRoute("index"), http.StatusSeeOther)
	} else {
		err = h(buf, req, ctx)
	}
	if err != nil {
		util.Log(err, LOG_HANDLERS)
	}
	if err = ctx.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Apply(w)
}

func RedirectHandler(dest string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, dest, http.StatusSeeOther)
	})
}

func FileHandler(origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sess, err := store.Get(req, "impendulo")
		if err != nil {
			util.Log(err, LOG_HANDLERS)
		}
		//Load our context from session
		ctx := LoadContext(sess)
		buf := new(HttpBuffer)
		err = CheckAccess(req.URL.Path, ctx, Permissions())
		var servePath string
		if err == nil {
			servePath, err = ServePath(req.URL, origin)
		}
		if err != nil {
			ctx.AddMessage(err.Error(), true)
			http.Redirect(buf, req, getRoute("index"), http.StatusSeeOther)
		} else {
			http.ServeFile(buf, req, servePath)
		}
		if err != nil {
			util.Log(err, LOG_HANDLERS)
		}
		if err = ctx.Save(req, buf); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buf.Apply(w)
	})
}

//getNav retrieves the navbar to display.
func getNav(ctx *Context) string {
	uname, err := ctx.Username()
	if err != nil {
		return "outnavbar"
	}
	u, err := db.User(uname)
	if err != nil {
		return "outnavbar"
	}
	switch u.Access {
	case user.STUDENT:
		return "studentnavbar"
	case user.TEACHER:
		return "teachernavbar"
	case user.ADMIN:
		return "adminnavbar"
	}
	return "outNavbar"
}

//downloadProject makes a project skeleton available for download.
func downloadProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	path, err := LoadSkeleton(req)
	if err != nil {
		ctx.AddMessage("Could not load project skeleton.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	} else {
		http.ServeFile(w, req, path)
	}
	return err
}
