//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package web

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
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
	other := html.EscapeString("target.args = 2,1,2\nsearch.multiple_errors = true\nasdas8823=quesq,322\n ")
	requests := []postHolder{
		postHolder{"/createjpf?projectid=" + id + "&addedlisteners=" +
			lPkg + "ExecTracker&addedlisteners=" + lPkg +
			"DeadlockAnalyzer&addedsearches=" + sPkg + "DFSearch&other=" + other, true},
		postHolder{"/createjpf?", false},
		postHolder{"/createjpf?projectid=" + id, true},
		postHolder{"/createjpf?projectid=" + id + "&addedlisteners=" +
			lPkg + "ExecTracker&addedlisteners=" + lPkg +
			"DeadlockAnalyzer&addedsearches=" + sPkg + "DFSearch", true},
		postHolder{"/createjpf?projectid=" + id + "&addedlisteners=" + lPkg, true},
	}
	testToolFunc(t, CreateJPF, requests)
}

func TestCreatePMD(t *testing.T) {
	id := bson.NewObjectId().Hex()
	requests := []postHolder{
		postHolder{"/createpmd?projectid=" + id + "&ruleid=java-basic&ruleid=java-controversial", true},
		postHolder{"/createpmd?projectid=" + id + "&ruleid=", true},
		postHolder{"/createpmd?projectid=" + id + "&ruleid=java-basic", true},
		postHolder{"/createpmd?projectid=", false},
		postHolder{"/createpmd?", false},
	}
	testToolFunc(t, CreatePMD, requests)
}

func testToolFunc(t *testing.T, f Poster, requests []postHolder) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	auth, enc, err := util.CookieKeys()
	if err != nil {
		t.Error(err)
	}
	store := sessions.NewCookieStore(auth, enc)
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

/*
func TestRunTool(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	p := project.New("Triangle", "user", "Java", []byte("data"))
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
	auth, enc, err := util.CookieKeys()
	store := sessions.NewCookieStore(auth, enc)
	go processing.Serve(processing.AMQP_URI, 5)
	req, err := http.NewRequest("POST", "/runtool?projectid="+p.Id.Hex()+
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
	processing.WaitIdle()
}
*/
