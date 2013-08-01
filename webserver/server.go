package webserver

import (
	"code.google.com/p/gorilla/pat"
	"github.com/godfried/impendulo/util"
	"net/http"
	"path/filepath"
)

var router *pat.Router

func init() {
	router = pat.New()
	router.Add("POST", "/login", handler(login))
	router.Add("POST", "/register", handler(register))
	router.Add("POST", "/logout", handler(logout))
	router.Add("POST", "/addproject", handler(addProject))
	router.Add("POST", "/addtest", handler(addTest))
	router.Add("POST", "/changeskeleton", handler(changeSkeleton))
	router.Add("POST", "/addjpf", handler(addJPF))
	router.Add("POST", "/submitarchive", handler(submitArchive))
	router.Add("POST", "/deleteproject", handler(deleteProject))
	router.Add("POST", "/deleteuser", handler(deleteUser))
	
	router.Add("GET", "/displayresult", handler(displayResult)).Name("displayresult")
	router.Add("GET", "/getfiles", handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissions", handler(getSubmissions)).Name("getsubmissions")
	generateViews(router)
	router.Add("GET", "/skeleton.zip", handler(downloadProject))
	router.Add("GET", "/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	router.Add("GET", "/", handler(loadView("homeView", "home"))).Name("index")
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
			util.Log(err)
		}
	}
	if err := http.ListenAndServeTLS(":8080", cert, key, router); err != nil {
		util.Log(err)
	}
}

func Run() {
	if err := http.ListenAndServe(":8080", router); err != nil {
		util.Log(err)
	}
}
