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
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"net/http"
	"testing"
)

type postHolder struct {
	url   string
	valid bool
}

func TestLogin(t *testing.T) {
	requests := []postHolder{
		postHolder{"/login?username=user&password=password", true},
		postHolder{"/login?username=a&password=password", false},
		postHolder{"/login?username=user&password=b", false},
		postHolder{"/password=password", false},
		postHolder{"/login?username=user&password=", false},
		postHolder{"/login?username=&password=password", false},
	}
	testUserFunc(t, Login, requests)
}

func TestRegister(t *testing.T) {
	requests := []postHolder{
		postHolder{"/login?username=user&password=password", false},
		postHolder{"/login?username=a&password=password", true},
		postHolder{"/login?username=user&password=b", false},
		postHolder{"/password=password", false},
		postHolder{"/login?username=user&password=", false},
		postHolder{"/login?username=&password=password", false},
	}
	testUserFunc(t, Register, requests)
}

func TestDeleteUser(t *testing.T) {
	requests := []postHolder{
		postHolder{"/deleteuser?username=user", true},
		postHolder{"/deleteuser?username=user", false},
		postHolder{"/deleteuser?username=", false},
		postHolder{"/password=password", false},
		postHolder{"/deleteuser?username=aaaa", false},
	}
	testUserFunc(t, DeleteUser, requests)
}

func testUserFunc(t *testing.T, f Poster, requests []postHolder) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	u := user.New("user", "password")
	err := db.Add(db.USERS, u)
	if err != nil {
		t.Error(err)
	}
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
		_, err = f(req, ctx)
		if ph.valid && err != nil {
			t.Error(err, ph.url)
		} else if !ph.valid && err == nil {
			t.Error(fmt.Errorf("Expected error for %s.", ph.url))
		}
	}
}

func createContext(req *http.Request, store sessions.Store) (ctx *Context, err error) {
	sess, err := store.Get(req, "test")
	if err == nil {
		ctx = LoadContext(sess)
	}
	return

}
