package db

import (
	"encoding/gob"
	"fmt"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
	"strings"
)

//GetCheckstyleResult retrieves a Result matching
//the given interface from the active database.
func GetCheckstyleResult(matcher, selector bson.M) (ret *checkstyle.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = checkstyle.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		dec := gob.NewDecoder(file)
		err = dec.Decode(&ret.Data)
	}
	return
}

//GetPMDResult retrieves a Result matching
//the given interface from the active database.
func GetPMDResult(matcher, selector bson.M) (ret *pmd.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = pmd.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		dec := gob.NewDecoder(file)
		err = dec.Decode(&ret.Data)
	}
	return
}

//GetFindbugsResult retrieves a Result matching
//the given interface from the active database.
func GetFindbugsResult(matcher, selector bson.M) (ret *findbugs.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = findbugs.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		dec := gob.NewDecoder(file)
		err = dec.Decode(&ret.Data)
	}
	return
}

//GetJPFResult retrieves a Result matching
//the given interface from the active database.
func GetJPFResult(matcher, selector bson.M) (ret *jpf.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = jpf.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		dec := gob.NewDecoder(file)
		err = dec.Decode(&ret.Data)
	}
	return
}

//GetJUnitResult retrieves aResult matching
//the given interface from the active database.
func GetJUnitResult(matcher, selector bson.M) (ret *junit.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		dec := gob.NewDecoder(file)
		err = dec.Decode(&ret.Data)
	}
	return
}

//GetJavacResult retrieves a JavacResult matching
//the given interface from the active database.
func GetJavacResult(matcher, selector bson.M) (ret *javac.Result, err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[project.NAME] = javac.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	} else if hasGridFile(ret, selector) {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.OpenId(ret.GetId())
		if err != nil {
			return
		}
		defer file.Close()
		var temp *javac.Result
		dec := gob.NewDecoder(file)
		err = dec.Decode(&temp)
		ret.Data = temp.Data
	}
	return
}

func hasGridFile(result tool.ToolResult, selector bson.M) bool {
	return (selector == nil || selector[project.DATA] == 1) && result.OnGridFS()
}

func getGridFile(session *mgo.Session, id interface{}) (holder interface{}, err error) {
	fs := session.DB("").GridFS("fs")
	var file *mgo.GridFile
	file, err = fs.OpenId(id)
	if err != nil {
		return
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	err = dec.Decode(holder)
	return
}

//GetToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func GetToolResult(name string, matcher, selector bson.M) (ret tool.ToolResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//GetDisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func GetDisplayResult(name string, matcher, selector bson.M) (ret tool.DisplayResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//GetDisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func GetGraphResult(name string, matcher, selector bson.M) (ret tool.GraphResult, err error) {
	switch name {
	case javac.NAME:
		ret, err = GetJavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = GetJPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = GetFindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = GetPMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = GetCheckstyleResult(matcher, selector)
	default:
		ret, err = GetJUnitResult(matcher, selector)
		if err != nil {
			err = fmt.Errorf("Unknown result %q.", name)
		}
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(res tool.ToolResult) (err error) {
	if res == nil {
		err = fmt.Errorf("Result is nil. In db/result.go.")
		return
	}
	matcher := bson.M{project.ID: res.GetFileId()}
	change := bson.M{SET: bson.M{project.RESULTS + "." + res.GetName(): res.GetId()}}
	err = Update(FILES, matcher, change)
	if err != nil {
		return
	}
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	if res.OnGridFS() {
		fs := session.DB("").GridFS("fs")
		var file *mgo.GridFile
		file, err = fs.Create("")
		if err != nil {
			return
		}
		defer file.Close()
		file.SetId(res.GetId())
		enc := gob.NewEncoder(file)
		err = enc.Encode(res.GetData())
		if err != nil {
			return
		}
		res.SetData(nil)
	}
	col := session.DB("").C(RESULTS)
	err = col.Insert(res)
	if err != nil {
		err = &DBAddError{res.GetName(), err}
	}
	return
}

//AddTimeoutResult adds a new TimeoutResult to the active database.
func AddTimeoutResult(fileId bson.ObjectId, name string) (err error) {
	matcher := bson.M{project.ID: fileId}
	change := bson.M{SET: bson.M{project.RESULTS + "." + name: tool.TIMEOUT}}
	err = Update(FILES, matcher, change)
	return
}

//AddNoResult adds a new NoResult to the active database.
func AddNoResult(fileId bson.ObjectId, name string) (err error) {
	matcher := bson.M{project.ID: fileId}
	change := bson.M{SET: bson.M{project.RESULTS + "." + name: tool.NORESULT}}
	err = Update(FILES, matcher, change)
	return
}

//GetAllResults retrieves all tool.DisplayResults matching
//the given file Id from the active database.
func GetGraphResults(fileId bson.ObjectId) (ret []tool.GraphResult, err error) {
	file, err := GetFile(bson.M{project.ID: fileId}, bson.M{project.RESULTS: 1})
	if err != nil {
		return
	}
	ret = make([]tool.GraphResult, 0, len(file.Results))
	for name, id := range file.Results {
		if _, ok := id.(bson.ObjectId); !ok {
			continue
		}
		res, err := GetGraphResult(name, bson.M{project.ID: id}, nil)
		if err != nil {
			err = nil
			continue
		}
		ret = append(ret, res)
	}
	return
}

//GetResultNames retrieves all result names for a given project.
func GetResultNames(projectId bson.ObjectId, nonTool bool) (ret []string, err error) {
	tests, err := GetTests(bson.M{project.PROJECT_ID: projectId},
		bson.M{project.NAME: 1})
	if err != nil {
		return
	}
	ret = make([]string, len(tests))
	for i, test := range tests {
		ret[i] = strings.Split(test.Name, ".")[0]
	}
	ret = append(ret, checkstyle.NAME, findbugs.NAME,
		javac.NAME, jpf.NAME, pmd.NAME)
	if nonTool {
		ret = append(ret, tool.CODE, diff.NAME, tool.SUMMARY)
	}
	sort.Strings(ret)
	return
}

//RemoveResultById removes a result matching
//the given id from the active database.
func RemoveResultById(id interface{}) (err error) {
	session, err := getSession()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.RemoveId(id)
	if err != nil {
		err = &DBRemoveError{"result", err, id}
	}
	return
}
