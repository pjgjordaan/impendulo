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
	"testing"
)

func TestPackagePath(t *testing.T) {
	targ := NewTarget("", "", ".c.a", "d/a/fa")
	if targ.PackagePath() != "d/a/fa/c/a" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "", "d/a/fa")
	if targ.PackagePath() != "d/a/fa" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "a.v.a", "")
	if targ.PackagePath() != "a/v/a" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "..", "/")
	if targ.PackagePath() != "/" {
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
}

func TestExecutable(t *testing.T) {
	targ := NewTarget("hello.c", "", "c.a.d", "")
	if targ.Executable() != "c.a.d.hello" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("hello.c.d.a", "", "", "")
	if targ.Executable() != "hello" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "", "", "")
	if targ.Executable() != "" {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "", "a.b.c", "")
	if targ.Executable() != "a.b.c." {
		t.Errorf("Invalid executable %s", targ.Executable())
	}
}
