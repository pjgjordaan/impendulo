package tool

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"strings"
	"time"
)

var timeLimit = 10 * time.Minute

func SetTimeout(minutes int) {
	timeLimit = time.Duration(minutes) * time.Minute
}

type Tool interface {
	GetName() string
	GetLang() string
	Run(fileId bson.ObjectId, target *TargetInfo) (Result, error)
}

type ExecResult struct {
	StdOut, StdErr []byte
	Err            error
}

func (this *ExecResult) HasStdErr() bool {
	return this.StdErr != nil && len(this.StdErr) > 0
}

func (this *ExecResult) HasStdOut() bool {
	return this.StdOut != nil && len(strings.TrimSpace(string(this.StdOut))) > 0
}

func RunCommand(args []string, stdin io.Reader) (res *ExecResult) {
	res = new(ExecResult)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Start()
	if err != nil {
		res.Err = &StartError{args, err}
		return
	}
	doneChan := make(chan error)
	go func() { doneChan <- cmd.Wait() }()
	select {
	case err := <-doneChan:
		if err != nil {
			res.Err = &EndError{args, err}
		}
	case <-time.After(timeLimit):
		cmd.Process.Kill()
		stdout.WriteString("\nCommand timed out.")
		stderr.WriteString("\nCommand timed out.")
		res.Err = &EndError{args, errors.New("Command timed out.")}
	}
	res.StdOut, res.StdErr = stdout.Bytes(), stderr.Bytes()
	return
}

type StartError struct {
	args []string
	err  error
}

func (this *StartError) Error() string {
	return fmt.Sprintf("Encountered startup error %q executing command %q", this.err, this.args)
}

type EndError struct {
	args []string
	err  error
}

func (this *EndError) Error() string {
	return fmt.Sprintf("Encountered end error %q executing command %q", this.err, this.args)
}
