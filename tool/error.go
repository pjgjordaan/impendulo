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

package tool

import (
	"fmt"
	"os"
)

type (
	//TimeoutError is an error used to indicate that a command timed out.
	TimeoutError struct {
		args []string
	}
	//StartError is an error used to indicate that a command failed to start.
	StartError struct {
		args []string
		err  error
	}

	//EndError is an error used to indicate that a command gave an error upon completion.
	EndError struct {
		args   []string
		err    error
		stdErr string
	}

	//CompileError is used to indicate that compilation failed.
	CompileError struct {
		name string
		msg  string
	}

	//XMLError represents an error which occurred when attempting to unmarshall XML into a struct.
	XMLError struct {
		err    error
		origin string
	}
)

//MemoryError checks whether an error is a memory error.
func MemoryError(e error) bool {
	pe, ok := e.(*os.PathError)
	return ok && pe.Err.Error() == "cannot allocate memory"
}

//AccessError checks whether an error is an access error.
func AccessError(e error) bool {
	ae, ok := e.(*os.PathError)
	return ok && ae.Err.Error() == "bad file descriptor"
}

//IsCompileError checks whether an error is a CompileError.
func IsCompileError(e error) bool {
	_, ok := e.(*CompileError)
	return ok
}

//IsTimeout checks whether an error is a timeout error.
func IsTimeout(e error) bool {
	if e != nil {
		_, ok := e.(*TimeoutError)
		return ok
	}
	return false
}

//IsEndError checks whether an error is an EndError.
func IsEndError(e error) bool {
	if e != nil {
		_, ok := e.(*EndError)
		return ok
	}
	return false
}

//Error
func (t *TimeoutError) Error() string {
	return fmt.Sprintf("command %q timed out", t.args)
}

//Error
func (s *StartError) Error() string {
	return fmt.Sprintf("start error %q executing command %q", s.err, s.args)
}

//Error
func (e *EndError) Error() string {
	return fmt.Sprintf("end error %q: %s executing command %q", e.err, e.stdErr, e.args)
}

//NewCompileError
func NewCompileError(name, msg string) *CompileError {
	return &CompileError{
		name: name,
		msg:  msg,
	}
}

//Error
func (c *CompileError) Error() string {
	return fmt.Sprintf("compile failure %q due to: %q.", c.name, c.msg)
}

//Error
func (x *XMLError) Error() string {
	return fmt.Sprintf("error %q parsing XML in %s", x.err, x.origin)
}

//NewXMLError
func NewXMLError(e error, origin string) *XMLError {
	return &XMLError{e, origin}
}
