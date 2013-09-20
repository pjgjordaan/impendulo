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
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
	"strconv"
	"testing"
)

func TestSetup(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
}

func TestCount(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	num := 100
	n, err := Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("Invalid count %q, should be %q", n, 0)
	}
	for i := 0; i < num; i++ {
		var s int = i / 10
		err = Add(PROJECTS, project.NewProject("name"+strconv.Itoa(s), "user", "lang", []byte{}))
		if err != nil {
			t.Error(err)
		}
	}
	n, err = Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != num {
		t.Errorf("Invalid count %q, should be %q", n, num)
	}
	n, err = Count(PROJECTS, bson.M{"name": "name0"})
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Errorf("Invalid count %q, should be %q", n, 10)
	}

}
