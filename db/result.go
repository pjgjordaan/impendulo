package db

import (
	"fmt"
	"github.com/godfried/impendulo/tool"
)

//GetResult retrieves a result matching the given interface from the active database.
func GetResult(matcher, selector interface{}) (*tool.Result, error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var ret *tool.Result
	err := c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		return nil, fmt.Errorf("Encountered error %q when retrieving result matching %q from db", err, matcher)
	}
	return ret, nil
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
