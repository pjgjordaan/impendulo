package tool

import (
	"bytes"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"os/exec"
)

type Tool interface {
	GetName() string
	GetLang() string
	Run(fileId bson.ObjectId, target *TargetInfo) (Result, error)
}

func RunCommand(args []string, stdin io.Reader) (stdout, stderr []byte, err error) {
	outBuff, errBuff, err := runCommand(args, stdin)
	stdout, stderr = outBuff.Bytes(), errBuff.Bytes()
	return
}

func runCommand(args []string, stdin io.Reader) (stdout, stderr bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Start()
	if err != nil {
		err = &StartError{args, err}
		return
	}
	err = cmd.Wait()
	if err != nil {
		err = &EndError{args,err}
	}
	return
}

type StartError struct{
	args []string
	err error
}

func (this *StartError) Error()string {
	return fmt.Sprintf("Encountered startup error %q executing command %q", this.err, this.args)
}

type EndError struct{
	args []string
	err error
}

func (this *EndError) Error()string{
	return fmt.Sprintf("Encountered end error %q executing command %q", this.err, this.args)
}