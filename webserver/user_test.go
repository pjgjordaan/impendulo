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
