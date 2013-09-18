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

package tool

import (
	"errors"
	"os"
	"testing"
)

func TestRunCommand(t *testing.T) {
	failCmd := []string{"chmod", "777"}
	execRes := RunCommand(failCmd, nil)
	if execRes.Err == nil {
		t.Error("Command should have failed")
	}
	succeedCmd := []string{"ls", "-a", "-l"}
	execRes = RunCommand(succeedCmd, nil)
	if execRes.Err != nil {
		t.Error(execRes.Err)
	}
	noCmd := []string{"lsa"}
	execRes = RunCommand(noCmd, nil)
	if _, ok := execRes.Err.(*StartError); !ok {
		t.Error("Command should not have started", execRes.Err)
	}
	SetTimeout(1)
	longCmd := []string{"sleep", "100"}
	execRes = RunCommand(longCmd, nil)
	if !IsTimeout(execRes.Err) {
		t.Error("Expected timeout, got ", execRes.Err)
	}
	SetTimeout(10)
}

func TestErrorChecks(t *testing.T) {
	memErr := &os.PathError{
		Err: errors.New("cannot allocate memory"),
	}
	accessErr := &os.PathError{
		Err: errors.New("bad file descriptor"),
	}
	if !MemoryError(memErr) {
		t.Error("Should be memory error.")
	}
	if !AccessError(accessErr) {
		t.Error("Should be access error.")
	}

}
