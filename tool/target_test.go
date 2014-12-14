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
	"github.com/godfried/impendulo/project"

	"testing"
)

func TestPackagePath(t *testing.T) {
	targ := NewTarget("", ".c.a", "d/a/fa", project.JAVA)
	if targ.PackagePath() != "d/a/fa/c/a" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "d/a/fa", project.JAVA)
	if targ.PackagePath() != "d/a/fa" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "a.v.a", "", project.JAVA)
	if targ.PackagePath() != "a/v/a" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "..", "/", project.JAVA)
	if targ.PackagePath() != "/" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
}

func TestExecutable(t *testing.T) {
	targ := NewTarget("hello.c", "c.a.d", "", project.JAVA)
	if targ.Executable() != "c.a.d.hello" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("hello.c.d.a", "", "", project.JAVA)
	if targ.Executable() != "hello.c.d" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "", "", project.JAVA)
	if targ.Executable() != "" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "a.b.c", "", project.JAVA)
	if targ.Executable() != "a.b.c." {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
}
