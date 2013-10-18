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
	//Setup the router.
	router = pat.New()
	router.Add("POST", "/login", Handler(login))
	router.Add("POST", "/register", Handler(register))
	router.Add("POST", "/logout", Handler(Logout))

	GeneratePosts(router)

	GenerateViews(router)

	router.Add("GET", "/configview", Handler(configView)).Name("configview")

	router.Add("GET", "/displaychart", Handler(showChart)).Name("displaychart")
	router.Add("GET", "/displayresult", Handler(showResult)).Name("displayresult")
	router.Add("GET", "/getfiles", Handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissionschart", Handler(getSubmissionsChart)).Name("getsubmissionschart")
	router.Add("GET", "/getsubmissions", Handler(getSubmissions)).Name("getsubmissions")
	router.Add("GET", "/skeleton.zip", Handler(downloadProject))
	router.Add("GET", "/static/", FileHandler(StaticDir()))
	router.Add("GET", "/static", RedirectHandler("/static/"))
	router.Add("GET", "/logs/", FileHandler(util.LogDir()))
	router.Add("GET", "/logs", RedirectHandler("/logs/"))
	router.Add("GET", "/", Handler(LoadView("homeView", "home"))).Name("index")
}

//StaticDir retrieves the directory containing all the static files for the web server.
func StaticDir() string {
	if staticDir != "" {
		return staticDir
	}
	staticDir = filepath.Join(util.InstallPath(), "static")
	return staticDir
}

//getRoute retrieves a route for a given name.
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
func Run(port string) {
	if Active() {
		return
	}
	setActive(true)
	defer setActive(false)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		util.Log(err)
	}
}

//Active is whether the server is currently running.
func Active() bool {
	return running
}

//setActive
func setActive(active bool) {
	running = active
}
