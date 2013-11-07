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

package gcc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"strconv"
)

type (
	Report struct {
		Id       bson.ObjectId    "_id"
		Type     tool.CompileType "type"
		Warnings int              "warnings"
		Errors   int              "errors"
		Data     []byte           "data"
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (ret *Report, err error) {
	data = bytes.TrimSpace(data)
	ret = &Report{
		Id:   id,
		Data: data,
	}
	ret.Warnings, err = calcCount("warning", data)
	if err != nil {
		return
	}
	ret.Errors, err = calcCount("error", data)
	if err != nil {
		return
	}
	if ret.Errors > 0 {
		ret.Type = tool.ERRORS
	} else if ret.Warnings > 0 {
		ret.Type = tool.WARNINGS
	} else {
		ret.Type = tool.SUCCESS
	}
	return
}

//Success tells us if compilation finished with no errors or warnings.
func (this *Report) Success() bool {
	return this.Type == tool.SUCCESS
}

//Header generates a string which briefly describes the compilation.
func (this *Report) Header() (header string) {
	if this.Success() {
		header = string(this.Data)
	} else if this.Type == tool.ERRORS {
		header = strconv.Itoa(this.Errors) + " Error"
		if this.Errors > 1 {
			header += "s"
		}
		if this.Warnings > 0 {
			header += " & " + strconv.Itoa(this.Warnings) + " Warning"
			if this.Warnings > 1 {
				header += "s"
			}
		}
	} else if this.Type == tool.WARNINGS {
		header = strconv.Itoa(this.Warnings) + " Warning"
		if this.Warnings > 1 {
			header += "s"
		}
	}
	return
}

func calcCount(tipe string, data []byte) (n int, err error) {
	prog := fmt.Sprintf("/^[^:]+:[0-9]+:[0-9]+: (%s):/ {e++} END {print e + 0} 1", tipe)
	args := []string{"awk", prog}
	stdIn := bytes.NewReader(data)
	res := tool.RunCommand(args, stdIn)
	if res.Err != nil {
		err = res.Err
		return
	} else if res.HasStdErr() {
		err = fmt.Errorf("Encountered awk error %s", string(res.StdErr))
		return
	}
	split := bytes.Split(bytes.TrimSpace(res.StdOut), []byte("\n"))
	if len(split) < 1 {
		err = fmt.Errorf("Awk output %s too short", string(res.StdOut))
		return
	}
	split = bytes.Split(bytes.TrimSpace(split[len(split)-1]), []byte(" "))
	if len(split) < 1 {
		err = fmt.Errorf("Awk output %s too short", string(res.StdOut))
		return
	}
	n, _ = strconv.Atoi(string(split[0]))
	return
}
