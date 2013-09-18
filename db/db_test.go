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
