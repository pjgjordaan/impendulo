//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package webserver

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
