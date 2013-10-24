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
	langs     []string
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
	if langs == nil {
		langs = []string{JAVA}
	}
	return langs
}

func Supported(lang string) bool {
	for _, l := range Langs() {
		if l == lang {
			return true
		}
	}
	return false
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
