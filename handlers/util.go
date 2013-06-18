package handlers

import (
	"html/template"
	"path/filepath"
	"os"
	"fmt"
	"io"
	"bufio"
	"strings"
)

var funcs = template.FuncMap{
	"reverse": reverse,
	"genHTML": genHTML,
}

func getPackage(r io.Reader)string{
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "package"{
			scanner.Scan()
			return strings.Split(scanner.Text(), ";")[0]
		}
	}
	return ""
}

func genHTML(name string, data []byte)(string, error){
	dir := filepath.Join("static","gen")
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while creating directory %q", err, dir)
	}
	name = filepath.Join(dir, name+".html")
	f, err := os.Create(name)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while creating file %q", err, name)
	}
	_, err = f.Write(data)
	if err != nil {
		return "", fmt.Errorf("Encountered error %q while writing data to %q", err, f)
	}
	return "/"+name, nil
}

var basicT = []string{filepath.Join("templates", "_base.html"),
	filepath.Join("templates", "index.html"),filepath.Join("templates","navbar.html"),
	filepath.Join("templates","sidebar.html"), filepath.Join("templates","messages.html")}

func T(names... string) *template.Template {
	t := template.New("_base.html").Funcs(funcs)
	all := make([]string, len(basicT) + len(names))
	end := copy(all, basicT)
	for i, name := range names{
		all[i+end] = filepath.Join("templates", name) 
	}
	t = template.Must(t.ParseFiles(all...))
	return t
}
