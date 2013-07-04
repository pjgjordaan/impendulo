package db

import (
	"github.com/godfried/impendulo/tool"
)

//GetResult retrieves a result matching the given interface from the active database.
func GetResult(matcher, selector interface{}) (ret *tool.Result, err error) {
	session := getSession()
	defer session.Close()
	c := session.DB("").C(RESULTS)
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &DBGetError{"result", err, matcher}
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(r *tool.Result)(err error){
	session := getSession()
	defer session.Close()
	col := session.DB("").C(RESULTS)
	err = col.Insert(r)
	if err != nil {
		err = &DBAddError{r, err}
	}
	return
}
