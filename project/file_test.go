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

package project

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	correctFile := &File{
		Name:    "File.java",
		Time:    int64(1256030454696),
		Type:    SRC,
		Package: "za.ac.sun.cs",
	}
	correct := "za_ac_sun_cs_File.java_1256030454696_123_c"
	incorrect := "za_ac_sun_cs_File.java_123_c"
	recvFile, err := ParseName(correct)
	if err != nil {
		t.Error(err)
	} else {
		correctFile.Id = recvFile.Id
		correctFile.SubId = recvFile.SubId
		if !reflect.DeepEqual(recvFile, correctFile) {
			t.Error(recvFile, "!=", correctFile)
		}
	}
	_, err = ParseName(incorrect)
	if err == nil {
		t.Error(incorrect, "is not a valid name")
	}

}
