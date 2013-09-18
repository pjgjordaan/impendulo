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
	u, err := db.User(uname)
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
	err = db.Add(db.USERS, u)
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
