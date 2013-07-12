package web

import (
	"fmt"
	"github.com/godfried/impendulo/util"
	"html/template"
	"path/filepath"
	"strings"
)

var funcs = template.FuncMap{
	"reverse":     reverse,
	"projectName": projectName,
	"date":        util.Date,
	"setBreaks":   setBreaks,
	"address":     address,
	"base": filepath.Base,
	"shortname": shortname,
}

func shortname(exec string) string {
	elements := strings.Split(exec, `.`)
	num := len(elements)
	if num < 2{
		return exec
	}
	return strings.Join(elements[num-2:], `.`)
}

func address(val interface{}) string {
	return fmt.Sprint(&val)
}

func setBreaks(val string) template.HTML {
	return template.HTML(strings.Replace(val, "\n", "<br>", -1))
}

var dir = filepath.Join("static", "templates")
var basicT = []string{filepath.Join(dir, "_base.html"), filepath.Join(dir, "index.html"), filepath.Join(dir, "messages.html")}

func T(names ...string) *template.Template {
	t := template.New("_base.html").Funcs(funcs)
	all := make([]string, len(basicT)+len(names))
	end := copy(all, basicT)
	for i, name := range names {
		all[i+end] = filepath.Join(dir, name)
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
