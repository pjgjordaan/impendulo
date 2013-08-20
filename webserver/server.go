package webserver

import (
	"code.google.com/p/gorilla/pat"
	"github.com/godfried/impendulo/util"
	"net/http"
	"path/filepath"
)

var router *pat.Router

const LOG_SERVER = "webserver/server.go"

func init() {
	router = pat.New()
	router.Add("POST", "/login", Handler(login))
	router.Add("POST", "/register", Handler(register))
	router.Add("POST", "/logout", Handler(logout))
	router.Add("POST", "/deleteproject", Handler(deleteProject))
	router.Add("POST", "/deleteuser", Handler(deleteUser))

	GeneratePosts(router)

	GenerateViews(router)

	router.Add("GET", "/displaygraph", Handler(displayGraph)).Name("displaygraph")
	router.Add("GET", "/displayresult", Handler(displayResult)).Name("displayresult")
	router.Add("GET", "/getfiles", Handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissions", Handler(getSubmissions)).Name("getsubmissions")
	router.Add("GET", "/skeleton.zip", Handler(downloadProject))
	router.Add("GET", "/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	router.Add("GET", "/", Handler(LoadView("homeView", "home"))).Name("index")
}

func getRoute(name string) string {
	u, err := router.GetRoute(name).URL()
	if err != nil {
		return "/"
	}
	return u.Path
}

func RunTLS() {
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

func Run() {
	if err := http.ListenAndServe(":8080", router); err != nil {
		util.Log(err)
	}
}
