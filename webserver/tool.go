package webserver

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

var templates = map[string]string{
	jpf.NAME:        "jpfConfig",
	pmd.NAME:        "pmdConfig",
	junit.NAME:      "junitConfig",
	findbugs.NAME:   "findbugsConfig",
	checkstyle.NAME: "checkstyleConfig",
	"none":          "noConfig",
}

func toolTemplate(tool string) string {
	return templates[tool]
}

func toolPermissions() map[string]int {
	return map[string]int{
		"createjpf":        1,
		"createpmd":        1,
		"createjunit":      1,
		"createfindbugs":   1,
		"createcheckstyle": 1,
	}
}

func toolPosters() map[string]Poster {
	return map[string]Poster{
		"createpmd":        CreatePMD,
		"createjpf":        CreateJPF,
		"createjunit":      CreateJUnit,
		"createfindbugs":   CreateFindbugs,
		"createcheckstyle": CreateCheckstyle,
	}
}

func tools() []string {
	return []string{jpf.NAME, junit.NAME, pmd.NAME, findbugs.NAME, checkstyle.NAME}
}

func CreateCheckstyle(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

func CreateFindbugs(req *http.Request, ctx *Context) (msg string, err error) {
	return
}

func CreateJUnit(req *http.Request, ctx *Context) (msg string, err error) {
	username, err := ctx.Username()
	if err != nil {
		msg = "Could not read user."
		return
	}
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		msg = "Could not read project."
		return
	}
	testName, testBytes, err := ReadFormFile(req, "test")
	if err != nil {
		msg = "Could not read JUnit file."
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
			msg = "Could not read data file."
			return
		}
	}
	//Read package name from file.
	pkg := util.GetPackage(bytes.NewReader(testBytes))
	test := junit.NewTest(projectId, testName, username,
		pkg, testBytes, dataBytes)
	err = db.AddTest(test)
	return
}

//AddJPF replaces a project's JPF configuration file.
func AddJPF(req *http.Request, ctx *Context) (msg string, err error) {
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
func CreateJPF(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	username, err := ctx.Username()
	if err != nil {
		return
	}
	vals := make(map[string][]string)
	listeners, err := GetStrings(req, "addedL")
	if err == nil {
		vals["listener"] = listeners
	}
	search, err := GetString(req, "addedS")
	if err == nil {
		vals["search.class"] = []string{search}
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

func CreatePMD(req *http.Request, ctx *Context) (msg string, err error) {
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

func RunTool(req *http.Request, ctx *Context) (msg string, err error) {
	projectId, err := util.ReadId(req.FormValue("project"))
	if err != nil {
		return
	}
	tool, err := GetString(req, "tool")
	if err != nil {
		return
	}
	submissions, err := db.GetSubmissions(bson.M{project.PROJECT_ID: projectId}, bson.M{project.ID: 1})
	if err != nil {
		return
	}
	var runAll bool
	if req.FormValue("runempty-check") == "true" {
		runAll = false
	} else {
		runAll = true
	}
	for _, submission := range submissions {
		files, err := db.GetFiles(bson.M{project.SUBID: submission.Id}, bson.M{project.DATA: 0})
		if err != nil {
			util.Log(err)
			continue
		}
		err = processing.StartSubmission(submission.Id)
		if err != nil {
			util.Log(err)
			continue
		}
		for _, file := range files {
			if resultId, ok := file.Results[tool]; ok && runAll {
				err = db.RemoveResultById(resultId)
				if err != nil {
					util.Log(resultId, err)
					continue
				}
				delete(file.Results, tool)
				change := bson.M{db.SET: bson.M{project.RESULTS: file.Results}}
				err = db.Update(db.FILES, bson.M{project.ID: file.Id}, change)
				if err != nil {
					util.Log(err)
					continue
				}
			}
			err = processing.AddFile(file)
			if err != nil {
				util.Log(err)
			}
		}
		err = processing.EndSubmission(submission.Id)
		if err != nil {
			util.Log(err)
		}
	}
	return
}

//GetResultData retrieves a DisplayResult for a given file and result name.
func GetResultData(resultName string, fileId bson.ObjectId) (res tool.DisplayResult, err error) {
	var file *project.File
	matcher := bson.M{project.ID: fileId}
	file, err = db.GetFile(matcher, nil)
	if err != nil {
		return
	}
	switch resultName {
	case tool.CODE:
		res = tool.NewCodeResult(file.Data)
	case diff.NAME:
		res = diff.NewDiffResult(file)
	case tool.SUMMARY:
		res = tool.NewSummaryResult()
		//Load summary for each available result.
		for name, resid := range file.Results {
			var currentRes tool.ToolResult
			currentRes, err = db.GetToolResult(name,
				bson.M{project.ID: resid}, nil)
			if err != nil {
				return
			}
			res.(*tool.SummaryResult).AddSummary(currentRes)
		}
	default:
		ival, ok := file.Results[resultName]
		if !ok {
			res = tool.NewErrorResult(
				fmt.Errorf("No result available for %v.", resultName))
			return
		}
		switch val := ival.(type) {
		case bson.ObjectId:
			//Retrieve result from the db.
			matcher = bson.M{project.ID: val}
			res, err = db.GetDisplayResult(resultName,
				matcher, nil)
		case string:
			//Error, so create new error result.
			switch val {
			case tool.TIMEOUT:
				res = new(tool.TimeoutResult)
			case tool.NORESULT:
				res = new(tool.NoResult)
			default:
				res = tool.NewErrorResult(
					fmt.Errorf("No result available for %v.", resultName))
			}
		default:
			res = tool.NewErrorResult(
				fmt.Errorf("No result available for %v.", resultName))
		}
	}
	return
}
