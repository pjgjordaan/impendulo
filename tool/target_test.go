package tool

import (
	"testing"
)

func TestPackagePath(t *testing.T) {
	targ := NewTarget("", "", ".c.a", "d/a/fa")
	if targ.PackagePath() != "d/a/fa/c/a"{
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "", "d/a/fa")
	if targ.PackagePath() != "d/a/fa"{
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "a.v.a", "")
	if targ.PackagePath() != "a/v/a"{
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
	targ = NewTarget("", "", "..", "/")
	if targ.PackagePath() != "/"{
		t.Errorf("Invalid package path %s", targ.PackagePath())
	}
}

func TestExecutable(t *testing.T){
	targ := NewTarget("hello.c", "", "c.a.d", "")
	if targ.Executable() != "c.a.d.hello"{
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("hello.c.d.a", "", "", "")
	if targ.Executable() != "hello"{
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "", "", "")
	if targ.Executable() != ""{
		t.Errorf("Invalid executable %s", targ.Executable())
	}
	targ = NewTarget("", "", "a.b.c", "")
	if targ.Executable() != "a.b.c."{
		t.Errorf("Invalid executable %s", targ.Executable())
	}
}