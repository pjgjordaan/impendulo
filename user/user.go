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

//Package user provides data structures and methods for interacting
//with users in the Impendulo system.
package user

import (
	"bufio"
	"fmt"
	"github.com/godfried/impendulo/util"
	"os"
	"strings"
)

type (
	Permission int

	//User represents a user within the Impendulo system.
	User struct {
		Name     string     "_id"
		Password string     "password"
		Salt     string     "salt"
		Access   Permission "access"
	}
)

const (
	//Permissions
	NONE Permission = iota
	STUDENT
	TEACHER
	ADMIN
	//struct db names
	ID     = "_id"
	PWORD  = "password"
	SALT   = "salt"
	ACCESS = "access"
)

var (
	perms []Permission
)

//String
func (this *User) String() string {
	return "Type: user.User; Name: " + this.Name
}

//New creates a new user with file submission permissions.
func New(uname, pword string) *User {
	hash, salt := util.Hash(pword)
	return &User{uname, hash, salt, STUDENT}
}

//Read reads user configurations from a file.
//It also sets up their passwords.
func Read(fname string) (users []*User, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(f)
	defaultSize := 1000
	users = make([]*User, 0, defaultSize)
	for scanner.Scan() {
		vals := strings.Split(scanner.Text(), ":")
		if len(vals) != 2 {
			err = fmt.Errorf("Line %s in config file not formatted correctly.", scanner.Text())
			return
		}
		uname := strings.TrimSpace(vals[0])
		pword := strings.TrimSpace(vals[1])
		users = append(users, New(uname, pword))
	}
	err = scanner.Err()
	return

}

func ValidPermission(perm int) bool {
	switch Permission(perm) {
	case NONE, STUDENT, TEACHER, ADMIN:
		return true
	default:
		return false
	}
}

func (this Permission) Name() string {
	switch this {
	case NONE:
		return "None"
	case STUDENT:
		return "Student"
	case TEACHER:
		return "Teacher"
	case ADMIN:
		return "Administrator"
	default:
		return "Unknown"
	}
}

func Permissions() []Permission {
	if perms == nil {
		perms = []Permission{NONE, STUDENT, TEACHER, ADMIN}
	}
	return perms
}
