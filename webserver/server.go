package webserver

import (
	"code.google.com/p/gorilla/pat"
	"fmt"
	"github.com/godfried/impendulo/util"
	"net/http"
	"path/filepath"
)

var router *pat.Router

func init() {
	router = pat.New()
	router.Add("POST", "/login", handler(login))
	router.Add("GET", "/displayresult", handler(displayResult)).Name("displayresult")
	router.Add("GET", "/getfiles", handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissions", handler(getSubmissions)).Name("getsubmissions")
	router.Add("GET", "/getprojects", handler(getProjects)).Name("getprojects")
	router.Add("GET", "/getusers", handler(getUsers)).Name("getusers")
	router.Add("GET", "/registerview", handler(registerView)).Name("registerview")
	router.Add("POST", "/register", handler(register))
	router.Add("POST", "/logout", handler(logout))
	router.Add("POST", "/addproject", handler(addProject))
	router.Add("POST", "/addtest", handler(addTest))
	router.Add("POST", "/changeskeleton", handler(changeSkeleton))
	router.Add("POST", "/addjpf", handler(addJPF))
	router.Add("POST", "/submitarchive", handler(submitArchive))
	router.Add("GET", "/projectdownloadview", handler(projectDownloadView)).Name("projectdownloadview")
	router.Add("GET", "/userdeleteview", handler(userDeleteView)).Name("userdeleteview")
	router.Add("GET", "/projectdeleteview", handler(projectDeleteView)).Name("projectdeleteview")
	router.Add("POST", "/deleteproject", handler(deleteProject))
	router.Add("POST", "/deleteuser", handler(deleteUser))
	router.Add("GET", "/skeleton.zip", handler(downloadProject))
	router.Add("GET", "/projectview", handler(projectView)).Name("projectview")
	router.Add("GET", "/skeletonview", handler(skeletonView)).Name("skeletonview")
	router.Add("GET", "/testview", handler(testView)).Name("testview")
	router.Add("GET", "/jpffileview", handler(jpfFileView)).Name("jpffileview")
	router.Add("GET", "/jpfconfigview", handler(jpfConfigView)).Name("jpfconfigview")
	router.Add("GET", "/homeview", handler(homeView)).Name("homeview")
	router.Add("GET", "/archiveview", handler(archiveView)).Name("archiveview")
	router.Add("GET", "/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	router.Add("GET", "/", handler(homeView)).Name("index")
}

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
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