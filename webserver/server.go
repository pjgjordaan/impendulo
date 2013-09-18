//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

//Package webserver provides a webserver which allows for: viewing
//of results; administration of submissions, projects and tools; user management;
package webserver

import (
	"code.google.com/p/gorilla/pat"
	"github.com/godfried/impendulo/util"
	"net/http"
	"path/filepath"
)

var (
	router    *pat.Router
	staticDir string
	running   bool
)

const (
	LOG_SERVER = "webserver/server.go"
)

func init() {
	router = pat.New()
	router.Add("POST", "/login", Handler(login))
	router.Add("POST", "/register", Handler(register))
	router.Add("POST", "/logout", Handler(logout))

	GeneratePosts(router)

	GenerateViews(router)

	router.Add("GET", "/configview", Handler(configView)).Name("configview")

	router.Add("GET", "/displaygraph", Handler(displayGraph)).Name("displaygraph")
	router.Add("GET", "/displayresult", Handler(displayResult)).Name("displayresult")
	router.Add("GET", "/getfiles", Handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissions", Handler(getSubmissions)).Name("getsubmissions")
	router.Add("GET", "/skeleton.zip", Handler(downloadProject))
	router.Add("GET", "/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(StaticDir()))))
	router.Add("GET", "/", Handler(LoadView("homeView", "home"))).Name("index")
}

//StaticDir
func StaticDir() string {
	if staticDir != "" {
		return staticDir
	}
	staticDir = filepath.Join(util.InstallPath(), "static")
	return staticDir
}

//getRoute
func getRoute(name string) string {
	u, err := router.GetRoute(name).URL()
	if err != nil {
		return "/"
	}
	return u.Path
}

//RunTLS
func RunTLS() {
	if Active() {
		return
	}
	setActive(true)
	defer setActive(false)
	cert := filepath.Join(util.BaseDir(), "cert.pem")
	key := filepath.Join(util.BaseDir(), "key.pem")
	if !util.Exists(cert) || !util.Exists(key) {
		err := util.GenCertificate(cert, key)
		if err != nil {
			util.Log(err, LOG_SERVER)
		}
	}
	if err := http.ListenAndServeTLS(":8080", cert, key, router); err != nil {
		util.Log(err, LOG_SERVER)
	}
}

//Run starts up the webserver if it is not currently running.
func Run() {
	if Active() {
		return
	}
	setActive(true)
	defer setActive(false)
	if err := http.ListenAndServe(":8080", router); err != nil {
		util.Log(err)
	}
}

//Active
func Active() bool {
	return running
}

//setActive
func setActive(active bool) {
	running = active
}
