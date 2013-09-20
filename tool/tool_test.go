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
