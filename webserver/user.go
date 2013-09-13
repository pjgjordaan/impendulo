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
	uname, pword, msg, err := getCredentials(req)
	if err != nil {
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
	uname, pword, msg, err := getCredentials(req)
	if err != nil {
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

//getCredentials
func getCredentials(req *http.Request) (uname, pword, msg string, err error) {
	uname, msg, err = getString(req, "username")
	if err != nil {
		return
	}
	pword, msg, err = getString(req, "password")
	return
}

//DeleteUser removes a user and all data associated with them from the system.
func DeleteUser(req *http.Request, ctx *Context) (msg string, err error) {
	uname, msg, err := getString(req, "username")
	if err != nil {
		return
	}
	err = db.RemoveUserById(uname)
	if err != nil {
		msg = fmt.Sprintf("Could not delete user %s.", uname)
	} else {
		msg = fmt.Sprintf("Successfully deleted user %s.", uname)
	}
	return
}
