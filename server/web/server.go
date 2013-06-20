package web

import (
	"code.google.com/p/gorilla/pat"
	"fmt"
	"net/http"
)

var router *pat.Router

func init() {
	router = pat.New()
	router.Add("POST", "/login", handler(login))
	router.Add("GET", "/getresults", handler(getResults)).Name("getresults")
	router.Add("GET", "/getfiles", handler(getFiles)).Name("getfiles")
	router.Add("GET", "/getsubmissions", handler(getSubmissions)).Name("getsubmissions")
	router.Add("GET", "/getprojects", handler(getProjects)).Name("getprojects")
	router.Add("GET", "/getusers", handler(getUsers)).Name("getusers")
	router.Add("GET", "/registerview", handler(registerView)).Name("registerview")
	router.Add("POST", "/register", handler(register))
	router.Add("POST", "/logout", handler(logout))
	router.Add("POST", "/addproject", handler(addProject))
	router.Add("POST", "/addtest", handler(addTest))
	router.Add("POST", "/submitarchive", handler(submitArchive))
	router.Add("GET", "/projectview", handler(projectView)).Name("projectview")
	router.Add("GET", "/testview", handler(testView)).Name("testview")
	router.Add("GET", "/homeview", handler(homeView)).Name("homeview")
	router.Add("GET", "/archiveview", handler(archiveView)).Name("archiveview")
	router.Add("GET", "/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	router.Add("GET", "/gen/", http.StripPrefix("/gen/", http.FileServer(http.Dir("gen/"))))
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

func Run() {
	if err := http.ListenAndServe(":"+"8080", router); err != nil {
		panic(err)
	}
}
