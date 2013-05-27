package db 

import(
	"github.com/godfried/cabanga/tool"
"fmt"
)

//GetTool retrieves a tool matching the given interface from the active database.
func GetTool(matcher interface{}) (*tool.Tool, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(TOOLS)
	var ret *tool.Tool
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving tool matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetTools retrieves tools matching the given interface from the active database.
func GetTools(matcher interface{}) ([]*tool.Tool, error) {
	session := getSession()
	defer session.Close()
	tcol := session.DB("").C(TOOLS)
	var ret []*tool.Tool
	err := tcol.Find(matcher).All(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving tools matching %q from db", err, matcher)
	}
	return ret, nil
}

//GetResult retrieves a result matching the given interface from the active database.
func GetResult(matcher interface{}) (*tool.Result, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var ret *tool.Result
	err := c.Find(matcher).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving result matching %q from db", err, matcher)
	}
	return ret, nil
}

//AddTool adds a new tool to the active database.
func AddTool(t *tool.Tool) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(TOOLS)
	err := col.Insert(t)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding tool %q to db", err, t)
	}
	return nil
}

//AddResult adds a new result to the active database.
func AddResult(r *tool.Result) error {
	session := getSession()
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err := col.Insert(r)
	if err != nil {
		return fmt.Errorf("Encountered error %q when adding result %q to db", err, r)
	}
	return nil
}
