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
		args []string
		err  error
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
func MemoryError(err error) bool {
	pErr, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return pErr.Err.Error() == "cannot allocate memory"
}

//AccessError checks whether an error is an access error.
func AccessError(err error) bool {
	pErr, ok := err.(*os.PathError)
	if !ok {
		return false
	}
	return pErr.Err.Error() == "bad file descriptor"
}

//IsCompileError checks whether an error is a CompileError.
func IsCompileError(err error) (ok bool) {
	_, ok = err.(*CompileError)
	return
}

//IsTimeout checks whether an error is a timeout error.
func IsTimeout(err error) (ok bool) {
	if err != nil {
		_, ok = err.(*TimeoutError)
	}
	return
}

//IsEndError checks whether an error is an EndError.
func IsEndError(err error) (ok bool) {
	if err != nil {
		_, ok = err.(*EndError)
	}
	return
}

//Error
func (this *TimeoutError) Error() string {
	return fmt.Sprintf("Command %q timed out.", this.args)
}

//Error
func (this *StartError) Error() string {
	return fmt.Sprintf("Encountered startup error %q executing command %q",
		this.err, this.args)
}

//Error
func (this *EndError) Error() string {
	return fmt.Sprintf("Encountered end error %q executing command %q",
		this.err, this.args)
}

//NewCompileError
func NewCompileError(name, msg string) *CompileError {
	return &CompileError{
		name: name,
		msg:  msg,
	}
}

//Error
func (this *CompileError) Error() string {
	return fmt.Sprintf("Could not compile %q due to: %q.", this.name, this.msg)
}

//Error
func (this *XMLError) Error() string {
	return fmt.Sprintf("Encountered error %q while parsing xml in %s.",
		this.err, this.origin)
}

//NewXMLError
func NewXMLError(err error, origin string) *XMLError {
	return &XMLError{err, origin}
}
