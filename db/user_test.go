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

package db

import (
	"github.com/godfried/impendulo/user"
	"reflect"
	"testing"
)

func TestUser(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	u := user.New("uname", "pword")
	err = Add(USERS, u)
	if err != nil {
		t.Error(err)
	}
	found, err := User("uname")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, found) {
		t.Error("Users not equivalent", u, found)
	}
}
