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
	"time"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/result"
	"labix.org/v2/mgo/bson"

	"strconv"
)

type (
	Report struct {
		Id       bson.ObjectId      `bson:"_id"`
		Type     result.CompileType `bson:"type"`
		Warnings int                `bson:"warnings"`
		Errors   int                `bson:"errors"`
		Data     []byte             `bson:"data"`
	}
)

func init() {
	gob.Register(new(Report))
}

//NewReport
func NewReport(id bson.ObjectId, data []byte) (*Report, error) {
	data = bytes.TrimSpace(data)
	wc, e := calcCount("warning", data)
	if e != nil {
		return nil, e
	}
	ec, e := calcCount("error", data)
	if e != nil {
		return nil, e
	}
	t := result.SUCCESS
	if ec > 0 {
		t = result.ERRORS
	} else if wc > 0 {
		t = result.WARNINGS
	}
	return &Report{
		Id:       id,
		Data:     data,
		Warnings: wc,
		Errors:   ec,
		Type:     t,
	}, nil
}

//Success tells us if compilation finished with no errors or warnings.
func (r *Report) Success() bool {
	return r.Type == result.SUCCESS
}

//Header generates a string which briefly describes the compilation.
func (r *Report) Header() (header string) {
	if r.Success() {
		return string(r.Data)
	}
	return combine(countString(r.Errors, "Error"), countString(r.Warnings, "Warning"))
}

func countString(c int, n string) string {
	switch c {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%d %s", c, n)
	default:
		return fmt.Sprintf("%d %ss", c, n)
	}

}

func combine(a, b string) string {
	if a == "" {
		return b
	} else if b == "" {
		return a
	}
	return fmt.Sprintf("%s & %s", a, b)
}

func calcCount(tipe string, data []byte) (int, error) {
	prog := fmt.Sprintf("/^[^:]+:[0-9]+:[0-9]+: (%s):/ {e++} END {print e + 0} 1", tipe)
	r, e := tool.RunCommand([]string{"awk", prog}, bytes.NewReader(data), 30*time.Second)
	if e != nil {
		return -1, e
	} else if r.HasStdErr() {
		return -1, fmt.Errorf("encountered awk error %s", string(r.StdErr))
	}
	sp := bytes.Split(bytes.TrimSpace(r.StdOut), []byte("\n"))
	if len(sp) < 1 {
		return -1, fmt.Errorf("awk output %s too short", string(r.StdOut))
	}
	sp = bytes.Split(bytes.TrimSpace(sp[len(sp)-1]), []byte(" "))
	if len(sp) < 1 {
		return -1, fmt.Errorf("awk output %s too short", string(r.StdOut))
	}
	return strconv.Atoi(string(sp[0]))
}
