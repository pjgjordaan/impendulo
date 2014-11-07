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
	Permission     int
	PermissionInfo struct {
		Access Permission
		Name   string
	}

	//U represents a user within the Impendulo system.
	U struct {
		Name     string     `bson:"_id"`
		Password string     `bson:"password"`
		Salt     string     `bson:"salt"`
		Access   Permission `bson:"access"`
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
func (u *U) String() string {
	return "Type: user.U; Name: " + u.Name + "; Permission: " + u.Access.Name()
}

//New creates a new user with file submission permissions.
func New(u, p string) *U {
	h, s := util.Hash(p)
	return &U{u, h, s, STUDENT}
}

//Read reads user configurations from a file.
//It also sets up their passwords.
func Read(n string) ([]*U, error) {
	f, e := os.Open(n)
	if e != nil {
		return nil, e
	}
	s := bufio.NewScanner(f)
	us := make([]*U, 0, 1000)
	i := 0
	for s.Scan() {
		vs := strings.Split(s.Text(), ":")
		if len(vs) != 2 {
			return nil, fmt.Errorf("line %d %s formatted incorrectly", i, s.Text())
		}
		us = append(us, New(strings.TrimSpace(vs[0]), strings.TrimSpace(vs[1])))
		i++
	}
	if e = s.Err(); e != nil {
		return nil, e
	}
	return us, nil
}

func ValidPermission(p int) bool {
	switch Permission(p) {
	case NONE, STUDENT, TEACHER, ADMIN:
		return true
	default:
		return false
	}
}

func (p Permission) Name() string {
	switch p {
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

func PermissionInfos() []*PermissionInfo {
	ps := Permissions()
	infos := make([]*PermissionInfo, len(ps))
	for i, p := range ps {
		infos[i] = &PermissionInfo{p, p.Name()}
	}
	return infos
}
