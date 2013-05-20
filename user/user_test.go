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
