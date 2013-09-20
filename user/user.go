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
