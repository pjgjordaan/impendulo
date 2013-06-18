package web

import (
	"net/http"
	"github.com/godfried/impendulo/context"
	"fmt"
)

func homeView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T("homeView.html", "noRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}

func projectView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	langs := []string{"Java"}
	return T("projectView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "p":true, "langs": langs})
}

func testView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("testView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "t":true})
}

func archiveView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("archiveView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "a":true})
}

func registerView(w http.ResponseWriter, req *http.Request, ctx *context.Context) (err error) {
	return T("registerView.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "r":true})
}

func getUsers(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	if ctx.Users == nil{
		err := ctx.LoadUsers()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("homeView.html", "userRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}

func getProjects(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	if ctx.Projects == nil{
		err := ctx.LoadProjects()
		if err != nil {
			ctx.AddError(err.Error())
			http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
			return nil
		}
	}
	return T("homeView.html", "projRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true})
}


func getSubmissions(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	tp := req.FormValue("type")
	idStr := req.FormValue("id")
	subs, err := ctx.Subs(tp, idStr)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "subRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "subRes": subs})
}

func getFiles(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	subId := req.FormValue("subid")
	fileRes, err := ctx.GetFiles(subId)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "fileRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "fileRes": fileRes})
}

func getResults(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	res, err := buildResults(req)
	if err != nil {
		ctx.AddError(err.Error())
		http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
		return nil
	}
	return T("homeView.html", "dispRes.html", getTabs(ctx)).Execute(w, map[string]interface{}{"ctx": ctx, "h":true, "res": res})
}

func login(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	uname, err := processLogin(req)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.Session.Values["user"] = uname
	}
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func register(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	uname, err := processRegistration(req)
	if err != nil{
		ctx.AddError(err.Error())
		http.Redirect(w, req, reverse("registerview"), http.StatusSeeOther)
	} else{
		ctx.AddSuccess(fmt.Sprintf("Successfully registered as %q.", uname))
		ctx.Session.Values["user"] = uname
		http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	}
	return nil
}

func logout(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	delete(ctx.Session.Values, "user")
	http.Redirect(w, req, reverse("index"), http.StatusSeeOther)
	return nil
}

func addTest(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processTest(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.AddSuccess("Successfully added test.")
	}
	http.Redirect(w, req, reverse("testview"), http.StatusSeeOther)
	return nil
}

func addProject(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processProject(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	} else{
		ctx.AddSuccess("Successfully added project.")
	}
	http.Redirect(w, req, reverse("projectview"), http.StatusSeeOther)
	return nil
}

func submitArchive(w http.ResponseWriter, req *http.Request, ctx *context.Context) error {
	err := processArchive(req, ctx)
	if err != nil{
		ctx.AddError(err.Error())
	}else{
		ctx.AddSuccess("Successfully submitted archive.")
	}
	http.Redirect(w, req, reverse("archiveview"), http.StatusSeeOther)
	return nil
}
