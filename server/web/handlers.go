package web

import (
	"github.com/godfried/impendulo/context"
	"net/http"
)

func homeView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "homeView.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func projectView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	langs := []string{"Java"}
	return T(getNav(ctx), "projectView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true, "langs": langs})
}

func testView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "testView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func jpfFileView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "jpfFileView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func jpfConfigView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "jpfConfigView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func archiveView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "archiveView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
}

func registerView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "registerView.html").Execute(w, map[string]interface{}{"ctx": ctx, "r": true})
}

func getUsers(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	return T(getNav(ctx), "userRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func getProjects(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	return T(getNav(ctx), "projRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true})
}

func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	subs, msg, err := retrieveSubmissions(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	var temp string
	if ctx.Browse.IsUser {
		temp = "userSubRes.html"
	} else {
		temp = "projectSubRes.html"
	}
	return T(getNav(ctx), temp).Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	names, msg, err := retrieveNames(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(getNav(ctx), "fileRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "names": names})
}

func getResults(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	files, msg, err := retrieveFiles(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	curFile, msg, err := getCurrentFile(req, files[0].Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	selected, msg, err := getSelected(req, len(files))
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	if selected == len(files)-1{
		return T(getNav(ctx), "singleDispRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "curFile": curFile, "files": files, "selected": selected})
	}
	nextFile, msg, err := getFile(files[selected+1].Id)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	} 
	return T(getNav(ctx), "doubleDispRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "curFile": curFile, "files": files, "selected": selected, "nextFile":nextFile})
}

func login(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doLogin).exec(req, ctx)
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return err
}

func register(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doRegister).exec(req, ctx)
	if err != nil {
		http.Redirect(w, req, reverse("registerview"), http.StatusSeeOther)
	} else {
		http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	}
	return err
}

func logout(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func addTest(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doTest).exec(req, ctx)
	http.Redirect(w, req, reverse("testview"), http.StatusSeeOther)
	return err
}

func addJPF(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doJPF).exec(req, ctx)
	http.Redirect(w, req, reverse("jpffileview"), http.StatusSeeOther)
	return err
}

func addProject(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doProject).exec(req, ctx)
	http.Redirect(w, req, reverse("projectview"), http.StatusSeeOther)
	return err
}

func submitArchive(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processor(doArchive).exec(req, ctx)
	http.Redirect(w, req, reverse("archiveview"), http.StatusSeeOther)
	return err
}
