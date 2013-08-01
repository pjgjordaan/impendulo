package webserver

import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"net/http"
	"strings"
	"code.google.com/p/gorilla/pat"
)

var store sessions.Store

func init() {
	store = sessions.NewCookieStore(util.CookieKeys())
}

func getNav(ctx *Context) string {
	if _, err := ctx.Username(); err != nil {
		return "outNavbar"
	}
	return "inNavbar"
}

type handler func(http.ResponseWriter, *http.Request, *Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		util.Log(err)
	}
	ctx := NewContext(sess)
	buf := new(HttpBuffer)
	err = h(buf, req, ctx)
	if err != nil {
		util.Log(err)
	}
	if err = ctx.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Apply(w)
}

var views = map[string]string{"homeView":"home", "testView": "submit", 
	"skeletonView": "submit", "jpfFileView": "submit", 
	"registerView": "register", "projectDownloadView": "download",
	"projectDeleteView": "delete", "userDeleteView": "delete",
	"userResult": "home", "projectResult": "home",
	"jpfConfigView": "submit", "archiveView": "submit",
	"projectView": "submit"}

func generateViews(router *pat.Router){
	for name, view := range views{
		handleFunc := loadView(name, view)
		lname := strings.ToLower(name)
		pattern := "/"+lname
		router.Add("GET", pattern, handler(handleFunc)).Name(lname)
	}
}

func loadView(name, view string) handler{
	return func(w http.ResponseWriter, req *http.Request, ctx *Context) error{
		ctx.Browse.View = views[name] 
		args := map[string]interface{}{"ctx": ctx}
		return T(getNav(ctx), name).Execute(w, args)
	}
}

func downloadProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	path, err := loadSkeleton(req)
	if err == nil {
		http.ServeFile(w, req, path)
	} else {
		ctx.AddMessage("Could not load project skeleton.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	}
	return err
}

func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	subs, err := retrieveSubmissions(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve submissions.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	var temp string
	if ctx.Browse.IsUser {
		temp = "userSubmissionResult"
	} else {
		temp = "projectSubmissionResult"
	}
	ctx.Browse.View = "home"
	return T(getNav(ctx), temp).Execute(w, map[string]interface{}{"ctx": ctx, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	names, err := retrieveNames(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve files.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.Browse.View = "home"
	return T(getNav(ctx), "fileResult").Execute(w, map[string]interface{}{"ctx": ctx, "names": names})
}

func displayResult(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	args, temps, err := loadArgs(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not retrieve results." , true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	} 
	return T(temps...).Execute(w, args)
}

func loadArgs(req *http.Request, ctx *Context)(args map[string]interface{}, temps []string, err error) {
	files, err := retrieveFiles(req, ctx)
	if err != nil{
		return
	}
	selected, err := getSelected(req, len(files)-1)
	if err != nil{
		return
	}
	curFile, err := getFile(files[selected].Id)
	if err != nil{
		return
	}
	projectId, err := util.ReadId(ctx.Browse.Pid)
	if err != nil{
		return
	}
	results, err := db.GetResultNames(projectId)
	if err != nil{
		return
	}
	ctx.SetResult(req)
	res, err := GetResultData(ctx.Browse.Result, curFile.Id)
	if err != nil{
		return
	}
	ctx.Browse.View = "home"
	args = map[string]interface{}{"ctx": ctx, "files": files, 
		"selected": selected, "resultName": res.GetName(), 
		"curFile": curFile, "curResult": res.GetData(), 
		"results": results}
	neighbour, ok := getNeighbour(req, len(files)-1)
	temps = []string{getNav(ctx), "fileInfo", res.Template(true)}
	if !ok {
		temps = append(temps, "singleResult")
		return
	}
	nextFile, err := getFile(files[neighbour].Id)
	if err != nil {
		return
	}
	res, err = GetResultData(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		return
	}
	args["nextFile"] = nextFile
	args["nextResult"] = res.GetData()
	args["neighbour"] = neighbour
	temps = append(temps, "doubleResult", res.Template(false))
	return
}

func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doLogin(req, ctx)
	if err != nil {
		ctx.AddMessage("Invalid username or password." , true)
	} 
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return err
}

func register(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doRegister(req, ctx)
	if err != nil {
		ctx.AddMessage("Invalid credentials." , true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	} else {
		ctx.AddMessage("Successfully registered." , false)
		http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	}
	return err
}

func logout(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, getRoute("index"), http.StatusSeeOther)
	return nil
}

func addTest(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doTest(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not add test." , true)
	} else{
		ctx.AddMessage("Successfully added test." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func addJPF(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doJPF(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not add jpf config file." , true)
	} else{
		ctx.AddMessage("Successfully added jpf config file." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func addProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doProject(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not add project." , true)
	} else{
		ctx.AddMessage("Successfully added project.", false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func changeSkeleton(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doSkeleton(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not change project skeleton." , true)
	} else{
		ctx.AddMessage("Successfully changed project skeleton." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func submitArchive(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doArchive(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not submit Intlola archive." , true)
	} else{
		ctx.AddMessage("Submission successful." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func deleteProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doDeleteProject(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not delete project." , true)
	} else{
		ctx.AddMessage("Successfully deleted project." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}

func deleteUser(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := doDeleteUser(req, ctx)
	if err != nil {
		ctx.AddMessage("Could not delete user." , true)
	} else{
		ctx.AddMessage("Successfully deleted user." , false)
	}
	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	return err
}