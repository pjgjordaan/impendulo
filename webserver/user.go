package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"net/http"
)

//Login signs a user into the web app.
func Login(req *http.Request, ctx *Context) (msg string, err error) {
	uname, pword, err := getCredentials(req)
	if err != nil {
		msg = "Could not read credentials."
		return
	}
	u, err := db.GetUserById(uname)
	if err != nil {
		msg = fmt.Sprintf("User %s not found.", uname)
		return
	}
	if !util.Validate(u.Password, u.Salt, pword) {
		err = fmt.Errorf("Invalid username or password.")
		msg = err.Error()
	} else {
		ctx.AddUser(uname)
		msg = "Logged in successfully."
	}
	return
}

//Register registers a new user with Impendulo.
func Register(req *http.Request, ctx *Context) (msg string, err error) {
	uname, pword, err := getCredentials(req)
	if err != nil {
		msg = "Could not read credentials."
		return
	}
	u := user.New(uname, pword)
	err = db.AddUser(u)
	if err != nil {
		msg = fmt.Sprintf("User %s already exists.", uname)
	} else {
		ctx.AddUser(uname)
		msg = "Registered successfully."
	}
	return
}

func getCredentials(req *http.Request) (uname, pword string, err error) {
	uname, err = GetString(req, "username")
	if err != nil {
		return
	}
	pword, err = GetString(req, "password")
	return
}

//DeleteUser removes a user and all data associated with them from the system.
func DeleteUser(req *http.Request, ctx *Context) (err error) {
	uname, err := GetString(req, "user")
	if err == nil {
		err = db.RemoveUserById(uname)
	}
	return
}
