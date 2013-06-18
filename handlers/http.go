package handlers

import (
	"net/http"
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/httpbuf"
	"github.com/godfried/impendulo/context"
	"fmt"
	"os"
)

var store sessions.Store
//var session *mgo.Session
var Router *pat.Router

func init(){
	store = sessions.NewCookieStore([]byte(os.Getenv("KEY")))
	Router = pat.New()
	Router.Add("POST", "/login", handler(login))
	Router.Add("GET", "/getresults", handler(getResults)).Name("getresults")
	Router.Add("GET", "/getfiles", handler(getFiles)).Name("getfiles")
	Router.Add("GET", "/getsubmissions", handler(getSubmissions)).Name("getsubmissions")
	Router.Add("GET", "/getprojects", handler(getProjects)).Name("getprojects")
	Router.Add("GET", "/getusers", handler(getUsers)).Name("getusers")
	Router.Add("GET", "/registerview", handler(registerView)).Name("registerview")
	Router.Add("POST", "/register", handler(register))
	Router.Add("POST", "/logout", handler(logout))
	Router.Add("POST", "/addproject", handler(addProject))
	Router.Add("POST", "/addtest", handler(addTest))	
	Router.Add("POST", "/submitarchive", handler(submitArchive))
	Router.Add("GET", "/projectview", handler(projectView)).Name("projectview")
	Router.Add("GET", "/testview", handler(testView)).Name("testview")	
	Router.Add("GET", "/homeview", handler(homeView)).Name("homeview")
	Router.Add("GET", "/archiveview", handler(archiveView)).Name("archiveview")
	Router.Add("GET","/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	Router.Add("GET","/gen/", http.StripPrefix("/gen/", http.FileServer(http.Dir("gen/"))))
	Router.Add("GET", "/", handler(homeView)).Name("index")
}

func reverse(name string, things ...interface{}) string {
	//convert the things to strings
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	//grab the route
	u, err := Router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}


type handler func(http.ResponseWriter, *http.Request, *context.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//create the context
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.NewContext(req, sess)
	//run the handler and grab the error, and report it
	buf := new(httpbuf.Buffer)
	err = h(buf, req, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//save the session
	if err = ctx.Session.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//apply the buffered response to the writer
	buf.Apply(w)
}

