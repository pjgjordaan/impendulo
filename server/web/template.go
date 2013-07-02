package web

import (
	"github.com/godfried/impendulo/util"
	"html/template"
	"path/filepath"
)

var funcs = template.FuncMap{
	"reverse":     reverse,
	"genHTML":     genHTML,
	"getResult":   getResult,
	"projectName": projectName,
}

func genHTML(name string, data []byte) string {
	val, err := util.GenHTML(filepath.Join("static", "gen"), name, data)
	if err != nil {
		util.Log(err)
	}
	return val
}

var basicT = []string{filepath.Join("templates", "_base.html"), filepath.Join("templates", "index.html"), filepath.Join("templates", "messages.html")}

func T(names ...string) *template.Template {
	t := template.New("_base.html").Funcs(funcs)
	all := make([]string, len(basicT)+len(names))
	end := copy(all, basicT)
	for i, name := range names {
		all[i+end] = filepath.Join("templates", name)
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
