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
