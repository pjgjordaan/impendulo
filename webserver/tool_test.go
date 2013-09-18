package webserver

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/util"
	"html"
	"labix.org/v2/mgo/bson"
	"net/http"
	"testing"
)

var (
	fileInfo = map[string]interface{}{
		project.TIME: util.CurMilis(),
		project.TYPE: project.SRC,
		project.PKG:  "triangle",
		project.NAME: "Triangle.java",
	}

	fileData = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)
)

func TestCreateJPF(t *testing.T) {
	id := bson.NewObjectId().Hex()
	lPkg := "gov.nasa.jpf.listener."
	sPkg := "gov.nasa.jpf.search."
	other := html.EscapeString("target.args = 2,1,2\nsearch.multiple_errors = true\nasdas8823\n  ")
	requests := []postHolder{
		postHolder{"/createjpf?project=" + id + "&addedL=" +
			lPkg + "ExecTracker&addedL=" + lPkg +
			"DeadlockAnalyzer&addedS=" + sPkg + "DFSearch&other=" + other, true},
		postHolder{"/createjpf?", false},
		postHolder{"/createjpf?project=" + id, true},
		postHolder{"/createjpf?project=" + id + "&addedL=" +
			lPkg + "ExecTracker&addedL=" + lPkg +
			"DeadlockAnalyzer&addedS=" + sPkg + "DFSearch", true},
		postHolder{"/createjpf?project=" + id + "&addedL=" + lPkg, true},
	}
	testToolFunc(t, CreateJPF, requests)
}

func TestCreatePMD(t *testing.T) {
	id := bson.NewObjectId().Hex()
	requests := []postHolder{
		postHolder{"/createpmd?project=" + id + "&ruleid=java-basic&ruleid=java-controversial", true},
		postHolder{"/createpmd?project=" + id + "&ruleid=", true},
		postHolder{"/createpmd?project=" + id + "&ruleid=java-basic", true},
		postHolder{"/createpmd?project=", false},
		postHolder{"/createpmd?", false},
	}
	testToolFunc(t, CreatePMD, requests)
}

func testToolFunc(t *testing.T, f Poster, requests []postHolder) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	store := sessions.NewCookieStore(util.CookieKeys())
	for _, ph := range requests {
		req, err := http.NewRequest("POST", ph.url, nil)
		if err != nil {
			t.Error(err)
		}
		ctx, err := createContext(req, store)
		if err != nil {
			t.Error(err)
		}
		ctx.AddUser("user")
		_, err = f(req, ctx)
		if ph.valid && err != nil {
			t.Error(err)
		} else if !ph.valid && err == nil {
			t.Error(fmt.Errorf("Expected error for %s.", ph.url))
		}
	}
}

func TestRunTool(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	p := project.NewProject("Triangle", "user", "Java", []byte("data"))
	err := db.Add(db.PROJECTS, p)
	if err != nil {
		t.Error(err)
	}
	s := project.NewSubmission(p.Id, "user", project.FILE_MODE, util.CurMilis())
	err = db.Add(db.SUBMISSIONS, s)
	if err != nil {
		t.Error(err)
	}
	f, err := project.NewFile(s.Id, fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = db.Add(db.FILES, f)
	if err != nil {
		t.Error(err)
	}
	store := sessions.NewCookieStore(util.CookieKeys())
	go processing.Serve(5)
	req, err := http.NewRequest("POST", "/runtool?project="+p.Id.Hex()+
		"&tool="+findbugs.NAME+"&runempty-check=true", nil)
	if err != nil {
		t.Error(err)
	}
	ctx, err := createContext(req, store)
	if err != nil {
		t.Error(err)
	}
	_, err = RunTool(req, ctx)
	if err != nil {
		t.Error(err)
	}
	processing.Shutdown()
}
