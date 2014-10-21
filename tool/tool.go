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
	"fmt"

	"github.com/godfried/impendulo/tool/result"

	"io"

	"labix.org/v2/mgo/bson"

	"os/exec"
	"strings"
	"time"
)

type (
	//T is an interface which represents various analysis tools used in Impendulo.
	T interface {
		//Name retrieves the Tool's name.
		Name() string
		//Lang retrieves the language which the Tool is used for.
		Lang() Language
		//Run runs the tool on a given file.
		Run(fileId bson.ObjectId, target *Target) (result.Tooler, error)
	}
	Compiler interface {
		//Name retrieves the Tool's name.
		Name() string
		//Lang retrieves the language which the Tool is used for.
		Lang() Language
		//Run runs the tool on a given file.
		Run(fileId bson.ObjectId, target *Target) (result.Tooler, error)
		AddCP(string)
	}

	//Result is the result of RunCommand.
	Result struct {
		StdOut, StdErr []byte
	}
	Language string
)

const (
	JAVA Language = "Java"
	C    Language = "C"
	//The maximum size in bytes that a ToolResult is allowed to have.
	MAX_SIZE = 16000000
)

var (
	langs []Language
)

//Langs returns the languages supported by Impendulo
func Langs() []Language {
	if langs == nil {
		langs = []Language{JAVA, C}
	}
	return langs
}

func Supported(l Language) bool {
	for _, c := range Langs() {
		if l == c {
			return true
		}
	}
	return false
}

//HasStdErr checks whether the ExecResult has standard error output.
func (r *Result) HasStdErr() bool {
	return r.StdErr != nil && len(r.StdErr) > 0
}

//HasStdOut checks whether the ExecResult has standard output.
func (r *Result) HasStdOut() bool {
	return r.StdOut != nil && len(strings.TrimSpace(string(r.StdOut))) > 0
}

func (r *Result) String() string {
	return fmt.Sprintf("StdOut: %s;\nStdErr: %s;\n", string(r.StdOut), string(r.StdErr))
}

//RunCommand executes a given command given by args and stdin. It terminates
//when the command finishes execution or times out. A Result containing the
//command's output is returned.
func RunCommand(args []string, stdin io.Reader, max time.Duration) (*Result, error) {
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = stdin
	var so, se bytes.Buffer
	c.Stdout, c.Stderr = &so, &se
	e := c.Start()
	for MemoryError(e) || AccessError(e) {
		e = c.Start()
	}
	if e != nil {
		return nil, &StartError{args, e}
	}
	d := make(chan error)
	go func() {
		d <- c.Wait()
	}()
	select {
	case <-time.After(max):
		c.Process.Kill()
		return nil, &TimeoutError{args}
	case e := <-d:
		if e != nil {
			e = &EndError{args, e, string(se.Bytes())}
		}
		return &Result{StdOut: so.Bytes(), StdErr: se.Bytes()}, e
	}
}
