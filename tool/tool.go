package tool

import (
	"bytes"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"time"
	"errors"
)

type Tool interface {
	GetName() string
	GetLang() string
	Run(fileId bson.ObjectId, target *TargetInfo) (Result, error)
}

func RunCommand(args []string, stdin io.Reader) (stdout, stderr []byte, err error) {
	res := runCommand(args, stdin)
	stdout, stderr, err = res.stdout.Bytes(), res.stderr.Bytes(), res.err
	return
}

type execResult struct{
	stdout, stderr bytes.Buffer
	err error
}

func runCommand(args []string, stdin io.Reader) (res *execResult) {
	res = new(execResult)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout, cmd.Stderr = &res.stdout, &res.stderr
	err := cmd.Start()
	if err != nil {
		res.err = &StartError{args, err}
		return
	}
	doneChan := make(chan error)
	go func(){doneChan <- cmd.Wait()}() 
	select {
	case err := <-doneChan:
		if err != nil {
			res.err = &EndError{args,err}
		}
	case <-time.After(2 * time.Minute):
		cmd.Process.Kill()
		res.stdout.WriteString("\nCommand timed out.")
		res.stderr.WriteString("\nCommand timed out.")
		res.err = &EndError{args,errors.New("Command timed out.")}
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