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

package project

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	correctFile := &File{
		Name:     "File.java",
		Time:     int64(1256030454696),
		Type:     SRC,
		Package:  "za.ac.sun.cs",
		Comments: []*Comment{},
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
