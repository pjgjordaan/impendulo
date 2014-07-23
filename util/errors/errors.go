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

package errors

import (
	"errors"
	"fmt"
)

type (
	//missing indicates that a key was not present in a map.
	missing struct {
		key string
	}

	//cast indicates that an interface{} could not be cast to
	//a certain type.
	cast struct {
		tipe  string
		value interface{}
	}

	//util is used to add more context to errors which occur in the util package.
	util struct {
		origin interface{}
		tipe   string
		err    error
	}
	Writer struct{}
)

var (
	GoPath = errors.New("GOPATH is not set.")
	Found  = errors.New("directory was found")
)

func (w *Writer) Write(p []byte) (int, error) {
	return -1, errors.New("ERROR")
}

func NewMissing(k string) error {
	return &missing{key: k}
}

func NewCast(t string, v interface{}) error {
	return &cast{tipe: t, value: v}
}

func NewUtil(o interface{}, t string, e error) error {
	return &util{origin: o, tipe: t, err: e}
}

func (m *missing) Error() string {
	return fmt.Sprintf("read error: %q.", m.key)
}

func (c *cast) Error() string {
	return fmt.Sprintf("casting error: %q to %q.", c.value, c.tipe)
}

func IsCast(e error) bool {
	_, ok := e.(*cast)
	return ok
}

//Error
func (u *util) Error() string {
	return fmt.Sprintf("error %q while %s %q.", u.err, u.tipe, u.origin)
}
