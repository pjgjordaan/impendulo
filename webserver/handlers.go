package webserver

import (
	"net/http"
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/db"
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

func homeView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "homeView").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func projectView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	langs := []string{"Java"}
	return T(getNav(ctx), "projectView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true, "langs": langs})
}

func testView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "testView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func skeletonView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "skeletonView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func jpfFileView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "jpfFileView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func jpfConfigView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "jpfConfigView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func archiveView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "archiveView").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func registerView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "registerView").Execute(w, map[string]interface{}{"ctx": ctx, "r": true})
}

func projectDownloadView(w http.ResponseWriter, req *http.Request, ctx *Context) (err error) {
	return T(getNav(ctx), "projectDownloadView").Execute(w, map[string]interface{}{"ctx": ctx, "p": true})
}

func getUsers(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	return T(getNav(ctx), "userResult").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func downloadProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	path, err := loadSkeleton(req)
	if err == nil{
		http.ServeFile(w, req, path)
	}else{
		ctx.AddMessage(err.Error(), true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
	}
	return err
 }

func getProjects(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	return T(getNav(ctx), "projectResult").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	subs, msg, err := retrieveSubmissions(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	var temp string
	if ctx.Browse.IsUser {
		temp = "userSubmissionResult"
	} else {
		temp = "projectSubmissionResult"
	}
	return T(getNav(ctx), temp).Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	names, msg, err := retrieveNames(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(getNav(ctx), "fileResult").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "names": names})
}

func displayResult(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	files, msg, err := retrieveFiles(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	selected, msg, err := getSelected(req, len(files)-1)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	curFile, msg, err := getFile(files[selected].Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	ctx.SetResult(req)
	res, msg, err := getResult(ctx.Browse.Result, curFile.Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	curTemp, curResult := res.TemplateArgs(true)
	projectId, err := ReadId(ctx.Browse.Pid)
	if err != nil {
		ctx.AddMessage("Could not retrieve project identifier.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	results, err := db.GetResultNames(projectId)
	if err != nil {
		ctx.AddMessage("Could not retrieve results.", true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	args := map[string]interface{}{"ctx": ctx, "h": true, "files": files, "selected": selected, "resultName": res.GetName(), "curFile": curFile, "curResult": curResult, "results": results}
	neighbour, _, err := getNeighbour(req, len(files)-1)
	if err != nil {
		return T(getNav(ctx), "singleResult", "fileInfo", curTemp).Execute(w, args)
	}
	nextFile, msg, err := getFile(files[neighbour].Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	res, msg, err = getResult(ctx.Browse.Result, nextFile.Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	nextTemp, nextResult := res.TemplateArgs(false)
	args["nextFile"] = nextFile
	args["nextResult"] = nextResult
	args["neighbour"] = neighbour
	return T(getNav(ctx), "doubleResult", "fileInfo", curTemp, nextTemp).Execute(w, args)
}

func login(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doLogin).exec(req, ctx)
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return err
}

func register(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doRegister).exec(req, ctx)
	if err != nil {
		http.Redirect(w, req, reverse("registerview"), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	}
	return err
}

func logout(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func addTest(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doTest).exec(req, ctx)
	http.Redirect(w, req, reverse("testview"), http.StatusSeeOther)
	return err
}

func addJPF(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doJPF).exec(req, ctx)
	http.Redirect(w, req, reverse("jpffileview"), http.StatusSeeOther)
	return err
}

func addProject(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doProject).exec(req, ctx)
	http.Redirect(w, req, reverse("projectview"), http.StatusSeeOther)
	return err
}

func changeSkeleton(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doSkeleton).exec(req, ctx)
	http.Redirect(w, req, reverse("skeletonview"), http.StatusSeeOther)
	return err
}

func submitArchive(w http.ResponseWriter, req *http.Request, ctx *Context) error {
	err := processor(doArchive).exec(req, ctx)
	http.Redirect(w, req, reverse("archiveview"), http.StatusSeeOther)
	return err
}
