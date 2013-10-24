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

package webserver

import (
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
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

//Logout logs a user out of the system.
func Logout(req *http.Request, ctx *Context) (string, error) {
	delete(ctx.Session.Values, "user")
	return "Successfully logged out.", nil
}

//EditUser
func EditUser(req *http.Request, ctx *Context) (msg string, err error) {
	oldName, err := GetString(req, "oldname")
	if err != nil {
		msg = "Could not read old username."
		return
	}
	newName, err := GetString(req, "newname")
	if err != nil {
		msg = "Could not read new username."
		return
	}
	access, err := GetInt(req, "access")
	if err != nil {
		msg = "Could not read user access level."
		return
	}
	if !user.ValidPermission(access) {
		err = fmt.Errorf("Invalid user access level %d.", access)
		msg = err.Error()
		return
	}
	if oldName != newName {
		err = db.RenameUser(oldName, newName)
		if err != nil {
			msg = fmt.Sprintf("Could not rename user %s to %s.", oldName, newName)
			return
		}
	}
	change := bson.M{db.SET: bson.M{user.ACCESS: access}}
	err = db.Update(db.USERS, bson.M{db.ID: newName}, change)
	if err != nil {
		msg = "Could not edit user."
	} else {
		msg = "Successfully edited user."
	}
	//Ugly hack should change this.
	current := req.Header.Get("Referer")
	current = current[:strings.LastIndex(current, "/")+1] + "editdbview?editing=User"
	req.Header.Set("Referer", current)
	return
}
