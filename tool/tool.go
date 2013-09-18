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

//Package tool provides interfaces which tools must implement in order to be accepted into the Impendulo tool suite.
//These interfaces specify how a tool is run; what result it returns; and how the result is displayed.
package tool

import (
	"bytes"
	"io"
	"labix.org/v2/mgo/bson"
	"os/exec"
	"strings"
	"time"
)

const (
	JAVA = "Java"
	//The maximum size in bytes that a ToolResult is allowed to have.
	MAX_SIZE = 16000000
)

var (
	timeLimit = 5 * time.Minute
)

type (
	//Tool is an interface which represents various analysis tools used in Impendulo.
	Tool interface {
		//Name retrieves the Tool's name.
		Name() string
		//Lang retrieves the language which the Tool is used for.
		Lang() string
		//Run runs the tool on a given file.
		Run(fileId bson.ObjectId, target *TargetInfo) (ToolResult, error)
	}

	//ExecResult is the result of RunCommand.
	ExecResult struct {
		StdOut, StdErr []byte
		Err            error
	}
)

//Langs returns the languages supported by Impendulo
func Langs() []string {
	return []string{JAVA}
}

//SetTimeout sets the maximum time for which the RunCommand function can run.
func SetTimeout(minutes int) {
	timeLimit = time.Duration(minutes) * time.Minute
}

//Timeout returns the current timeout setting.
func Timeout() int {
	return int(timeLimit)
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
	for MemoryError(err) || AccessError(err) {
		err = cmd.Start()
	}
	if err != nil {
		res.Err = &StartError{args, err}
		return
	}
	doneChan := make(chan error)
	go func() {
		doneChan <- cmd.Wait()
	}()
	select {
	case err := <-doneChan:
		if err != nil {
			res.Err = &EndError{args, err}
		}
		res.StdOut, res.StdErr = stdout.Bytes(), stderr.Bytes()
	case <-time.After(timeLimit):
		cmd.Process.Kill()
		res.Err = &TimeoutError{args}
	}
	return
}
