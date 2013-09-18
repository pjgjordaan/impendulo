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

package user

import (
	"testing"
)

func TestHasAccess(t *testing.T) {
	u := &User{"uname", "pword", "salt", F_SUB}
	if !u.HasAccess(F_SUB) || u.HasAccess(NONE) || u.HasAccess(U_SUB) || u.HasAccess(T_SUB) {
		t.Error("Incorrect privilege")
	}
	u = &User{"uname", "pword", "salt", T_SUB}
	if !u.HasAccess(T_SUB) || u.HasAccess(NONE) || u.HasAccess(U_SUB) || u.HasAccess(F_SUB) {
		t.Error("Incorrect privilege")
	}
	u = &User{"uname", "pword", "salt", U_SUB}
	if !u.HasAccess(U_SUB) || u.HasAccess(NONE) || u.HasAccess(T_SUB) || u.HasAccess(F_SUB) {
		t.Error("Incorrect privilege")
	}
}

func CheckSubmit(t *testing.T) {
	u := &User{"uname", "pword", "salt", F_SUB}
	if !u.CheckSubmit(SINGLE) || u.CheckSubmit(TEST) || u.CheckSubmit(UPDATE) {
		t.Error("Incorrect mode")
	}
	u = &User{"uname", "pword", "salt", T_SUB}
	if u.CheckSubmit(SINGLE) || !u.CheckSubmit(TEST) || u.CheckSubmit(UPDATE) {
		t.Error("Incorrect mode")
	}
	u = &User{"uname", "pword", "salt", U_SUB}
	if u.CheckSubmit(SINGLE) || u.CheckSubmit(TEST) || !u.CheckSubmit(UPDATE) {
		t.Error("Incorrect mode")
	}
}
