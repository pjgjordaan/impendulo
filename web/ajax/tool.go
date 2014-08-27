package ajax

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web/webutil"

	"net/http"
)

func JPFListeners(r *http.Request) ([]byte, error) {
	l, e := jpf.Listeners()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"listeners": l})
}

func JPFSearches(r *http.Request) ([]byte, error) {
	s, e := jpf.Searches()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"searches": s})
}

func PMDRules(r *http.Request) ([]byte, error) {
	rs, e := pmd.RuleSet()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"rules": rs})
}

//Tools loads a list of available tools for a given project.
func Tools(r *http.Request) ([]byte, error) {
	pid, e := webutil.Id(r, "project-id")
	if e != nil {
		return nil, e
	}
	t, e := db.ProjectTools(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tools": t})
}
