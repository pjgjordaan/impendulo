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

const (
	//Permissions
	NONE = iota
	F_SUB
	T_SUB
	FT_SUB
	U_SUB
	UF_SUB
	UT_SUB
	ALL_SUB
	//Permission names
	SINGLE  = "file_remote"
	ARCHIVE = "archive_remote"
	TEST    = "archive_test"
	UPDATE  = "update"
	//struct db names
	ID     = "_id"
	PWORD  = "password"
	SALT   = "salt"
	ACCESS = "access"
)

type (
	//User represents a user within the Impendulo system.
	User struct {
		Name     string "_id"
		Password string "password"
		Salt     string "salt"
		Access   int    "access"
	}
)

//String
func (this *User) String() string {
	return "Type: user.User; Name: " + this.Name
}

//HasAccess checks whether a user has the required access level.
func (this *User) HasAccess(access int) bool {
	switch access {
	case NONE:
		return this.Access == NONE
	case F_SUB:
		return util.EqualsOne(this.Access, F_SUB, FT_SUB, UF_SUB, ALL_SUB)
	case T_SUB:
		return util.EqualsOne(this.Access, T_SUB, FT_SUB, UT_SUB, ALL_SUB)
	case U_SUB:
		return util.EqualsOne(this.Access, U_SUB, UF_SUB, UT_SUB, ALL_SUB)
	}
	return false
}

//CheckSubmit checks whether the user may provide the requested submission.
func (this *User) CheckSubmit(mode string) bool {
	if mode == SINGLE || mode == ARCHIVE {
		return this.HasAccess(F_SUB)
	} else if mode == TEST {
		return this.HasAccess(T_SUB)
	} else if mode == UPDATE {
		return this.HasAccess(U_SUB)
	}
	return false
}

//New creates a new user with file submission permissions.
func New(uname, pword string) *User {
	hash, salt := util.Hash(pword)
	return &User{uname, hash, salt, F_SUB}
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
