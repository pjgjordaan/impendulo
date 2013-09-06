package webserver

import (
	"bytes"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"net/http"
	"strings"
)

var templates = map[string]string{
	jpf.NAME:      "jpfConfig",
	pmd.NAME:      "pmdConfig",
	junit.NAME:    "junitConfig",
	findbugs.NAME: "findbugsConfig",
	"none":        "noConfig",
}

func toolTemplate(tool string) string {
	return templates[tool]
}

func toolPermissions() map[string]int {
	return map[string]int{
		"createjpf":      1,
		"createpmd":      1,
		"createjunit":    1,
		"createfindbugs": 1,
	}
}

func toolPostFuncs() map[string]PostFunc {
	return map[string]PostFunc{
		"createpmd":      CreatePMD,
		"createjpf":      CreateJPF,
		"createjunit":    CreateJUnit,
		"createfindbugs": CreateFindbugs,
	}
}

func tools() []string {
	return []string{jpf.NAME, junit.NAME, pmd.NAME, findbugs.NAME}
}

//AddTest adds a new test to a project.
func CreateFindbugs(req *http.Request, ctx *Context) (err error) {
	return nil
}

//AddTest adds a new test to a project.
func CreateJUnit(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	testName, testBytes, err := ReadFormFile(req, "test")
	if err != nil {
		return
	}
	hasData := req.FormValue("data-check")
	var dataBytes []byte
	if hasData == "" {
		dataBytes = make([]byte, 0)
	} else if hasData == "true" {
		//Read data files if provided.
		_, dataBytes, err = ReadFormFile(req, "data")
		if err != nil {
			return
		}
	}
	//Read package name from file.
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	username, err := ctx.Username()
	if err != nil {
		return
	}
	test := project.NewTest(projectId, testName, username,
		pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	return
}

//AddJPF replaces a project's JPF configuration file.
func AddJPF(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	_, data, err := ReadFormFile(req, "jpf")
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	jpfConfig := jpf.NewConfig(projectId, username, data)
	err = db.AddJPF(jpfConfig)
	return
}

//CreateJPF replaces a project's JPF configuration file.
func CreateJPF(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	listeners, err := GetStrings(req, "addedL")
	if err != nil {
		return
	}
	search, err := GetString(req, "addedS")
	if err != nil {
		return
	}
	vals := map[string][]string{
		"search.class": []string{search},
		"listener":     listeners,
	}
	other, err := GetString(req, "other")
	if err == nil {
		props := readProperties(other)
		for k, v := range props {
			vals[k] = v
		}
	}
	data, err := jpf.JPFBytes(vals)
	if err != nil {
		return
	}
	jpfConfig := jpf.NewConfig(projectId, username, data)
	err = db.AddJPF(jpfConfig)
	return
}

func readProperties(raw string) (props map[string][]string) {
	props = make(map[string][]string)
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		params := strings.Split(util.RemoveEmpty(line), "=")
		if len(params) == 2 {
			key, val := params[0], params[1]
			if len(key) > 0 && len(val) > 0 && jpf.Allowed(key) {
				split := strings.Split(val, ",")
				vals := make([]string, 0, len(split))
				for _, v := range split {
					if v != "" {
						vals = append(vals, v)
					}
				}
				if v, ok := props[key]; ok {
					props[key] = append(v, vals...)
				} else {
					props[key] = vals
				}
			}
		}
	}
	return
}

func CreatePMD(req *http.Request, ctx *Context) (err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	rules, err := GetStrings(req, "ruleid")
	if err != nil {
		return
	}
	pmdRules := pmd.NewRules(projectId, rules)
	err = db.AddPMD(pmdRules)
	return
}
