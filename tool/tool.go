package tool

import (
	"bytes"
	"fmt"
	"io"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"strings"
	"time"
)

var timeLimit = 10 * time.Minute

//SetTimeout sets the maximum time for which the RunCommand function can run.
func SetTimeout(minutes int) {
	timeLimit = time.Duration(minutes) * time.Minute
}

//Tool is an interface which represents various analysis tools used in Impendulo.
type Tool interface {
	//GetName retrieves the Tool's name.
	GetName() string
	//GetLang retrieves the language which the Tool is used for.
	GetLang() string
	//Run runs the tool on a given file.
	Run(fileId bson.ObjectId, target *TargetInfo) (ToolResult, error)
}

//ExecResult is the result of RunCommand.
type ExecResult struct {
	StdOut, StdErr []byte
	Err            error
}

//HasStdErr checks whether the ExecResult has standard error output.
func (this *ExecResult) HasStdErr() bool {
	return this.StdErr != nil && len(this.StdErr) > 0
}

//HasStdOut checks whether the ExecResult has standard output.
func (this *ExecResult) HasStdOut() bool {
	return this.StdOut != nil &&
		len(strings.TrimSpace(string(this.StdOut))) > 0
}

func MemoryError(err error)bool{
	pErr, ok := err.(*os.PathError); 
	if !ok {
		return false
	}
	return pErr.Err.Error() == "cannot allocate memory"
}

//RunCommand executes a given command given by args and stdin. It terminates
//when the command finishes execution or times out. An ExecResult containing the
//command's output is returned.
func RunCommand(args []string, stdin io.Reader) (res *ExecResult) {
	res = new(ExecResult)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Start()
	for MemoryError(err) {
		err = cmd.Start()
	} 
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
		res.StdOut, res.StdErr = stdout.Bytes(), stderr.Bytes()
	case <-time.After(timeLimit):
		cmd.Process.Kill()
		res.Err = &TimeOutError{args}
	}
	return
}

//TimeOutError is an error used to indicate that a command timed out.
type TimeOutError struct {
	args []string
}

func (this *TimeOutError) Error() string {
	return fmt.Sprintf("Command %q timed out.", this.args)
}

func IsTimeOut(err error) (ok bool) {
	_, ok = err.(*TimeOutError)
	return
}

//StartError is an error used to indicate that a command failed to start.
type StartError struct {
	args []string
	err  error
}

func (this *StartError) Error() string {
	return fmt.Sprintf("Encountered startup error %q executing command %q",
		this.err, this.args)
}

//EndError is an error used to indicate that a command gave an error upon completion.
type EndError struct {
	args []string
	err  error
}

func (this *EndError) Error() string {
	return fmt.Sprintf("Encountered end error %q executing command %q",
		this.err, this.args)
}
