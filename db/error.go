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

package db

import (
	"fmt"
)

type (
	//GetError represents errors encountered when retrieving data.
	GetError struct {
		tipe    string
		err     error
		matcher interface{}
	}

	//AddError represents errors encountered when adding data.
	AddError struct {
		msg string
		err error
	}

	//RemoveError represents errors encountered when removing data.
	RemoveError struct {
		tipe    string
		err     error
		matcher interface{}
	}
)

func (g *GetError) Error() string {
	return fmt.Sprintf("error %q  retrieving %q matching %q", g.err, g.tipe, g.matcher)
}

//Error
func (a *AddError) Error() string {
	return fmt.Sprintf("error %q adding %q", a.err, a.msg)
}

//Error
func (r *RemoveError) Error() string {
	return fmt.Sprintf("error %q removing %q matching %q", r.err, r.tipe, r.matcher)
}
