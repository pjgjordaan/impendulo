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
 
func jpfView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T(getNav(ctx), "jpfView.html").Execute(w, map[string]interface{}{"ctx": ctx, "s": true})
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
	if ctx.Browse.IsUser{
		temp = "userSubRes.html"
	} else{
		temp = "projectSubRes.html"
	}
	return T(getNav(ctx), temp).Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	fileRes, msg, err := retrieveFiles(req, ctx)
	if err != nil {
		ctx.AddMessage(msg, true)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(getNav(ctx), "fileRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "fileRes": fileRes})
}

func getResults(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	res, msg, err := buildResults(req)
	if err != nil {
		ctx.AddMessage(msg, err != nil)
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return err
	}
	return T(getNav(ctx), "dispRes.html").Execute(w, map[string]interface{}{"ctx": ctx, "h": true, "res": res})
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
	http.Redirect(w, req, reverse("jpfview"), http.StatusSeeOther)
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
