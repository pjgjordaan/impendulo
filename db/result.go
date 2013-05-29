package db 

import(
	"github.com/godfried/cabanga/tool"
"fmt"
)

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
