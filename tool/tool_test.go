package tool

import (
	"os"
	"testing"
	"errors"
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
	SetTimeout(0)
	longCmd := []string{"sleep", "100"}
	execRes = RunCommand(longCmd, nil)
	if !IsTimeOut(execRes.Err) {
		t.Error("Expected timeout, got ", execRes.Err)
	}
	SetTimeout(10)
}

func TestErrorChecks(t *testing.T){
	memErr := &os.PathError{
		Err:errors.New("cannot allocate memory"),
	}
	accessErr := &os.PathError{
		Err:errors.New("bad file descriptor"),
	}
	if !MemoryError(memErr){
		t.Error("Should be memory error.")
	}
	if !AccessError(accessErr){
		t.Error("Should be access error.")
	}
	
}