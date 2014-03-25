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
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, e := store.Get(r, "impendulo")
	if e != nil {
		util.Log(e, LOG_HANDLERS)
	}
	//Load our context from session
	c := LoadContext(s)
	b := new(HttpBuffer)
	e = CheckAccess(r.URL.Path, c, Permissions())
	if e != nil {
		c.AddMessage(e.Error(), true)
		http.Redirect(b, r, getRoute("index"), http.StatusSeeOther)
	} else {
		e = h(b, r, c)
	}
	if e != nil {
		util.Log(e, LOG_HANDLERS)
	}
	if e = c.Save(r, b); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
	b.Apply(w)
}

func RedirectHandler(dest string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, dest, http.StatusSeeOther)
	})
}

func FileHandler(origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, e := store.Get(r, "impendulo")
		if e != nil {
			util.Log(e, LOG_HANDLERS)
		}
		//Load our context from session
		c := LoadContext(s)
		b := new(HttpBuffer)
		e = CheckAccess(r.URL.Path, c, Permissions())
		var p string
		if e == nil {
			p, e = ServePath(r.URL, origin)
		}
		if e != nil {
			c.AddMessage(e.Error(), true)
			http.Redirect(b, r, getRoute("index"), http.StatusSeeOther)
		} else {
			http.ServeFile(b, r, p)
		}
		if e != nil {
			util.Log(e, LOG_HANDLERS)
		}
		if e = c.Save(r, b); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		b.Apply(w)
	})
}

//getNav retrieves the navbar to display.
func getNav(c *Context) string {
	n, e := c.Username()
	if e != nil {
		return "outnavbar"
	}
	u, e := db.User(n)
	if e != nil {
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
