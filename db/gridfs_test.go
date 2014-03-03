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
	"bytes"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

func TestResultGridFS(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	file, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = Add(FILES, file)
	if err != nil {
		t.Error(err)
	}
	results := []tool.ToolResult{
		checkstyleResult(file.Id, true),
		findbugsResult(file.Id, true),
		javacResult(file.Id, true),
	}
	for _, res := range results {
		err = AddResult(res, res.GetName())
		if err != nil {
			t.Error(err)
		}
		matcher := bson.M{"_id": res.GetId()}
		dbRes, err := ToolResult(res.GetName(), matcher, nil)
		if err != nil {
			t.Error(err)
		}
		res.SetReport(report(res.GetName(), res.GetId()))
		if !reflect.DeepEqual(res, dbRes) {
			t.Error("Results not equivalent", dbRes)
		}
	}
}

func TestGridFS(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	id := bson.NewObjectId()
	err = AddGridFile(id, grandPrix)
	if err != nil {
		t.Error(err)
	}
	var ret []byte
	err = GridFile(id, &ret)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(grandPrix, ret) {
		t.Error("Data not equal.")
	}
}

var grandPrix = []byte(`
Fast love,
Heart breaker,
I'll take you
In a tail... race.

Is this wonder land
My fire,
My passion
And mistake?

(Run and see,
Keep runnin' now...
Keep runnin' now,
I run and see)
I run and see
(Keep runnin' now)
Keep runnin' now,
(Run and see.)

I run and see,
(Run and see)
Keep runnin' now,
(Runnin' now.)

All because I wouldn't save your mind,
All because I wouldn't save...
All because I let you lose your life
When you lose your own way.

(I run and see)
It's so sad
(Keep runnin' now)
I will take her
(Run and see)
'Cause you're lost in...??
(Keep runnin' now)
Keep runnin' now
(I run and see.)

I can't blame you
(Keep runnin' now)
For getting me here,
Can't get you
(Run and see)
Out my head, yeah
(Runnin' now.)

Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now
Keep' runnin' now, keep' runnin' now...

Come close,
Open the door,
Draw the curtains
On the stairs...

Half dreamer,
Cream in the eye,
Say goodbye
To the child... there.

{I run and see,
Keep runnin' now,
I run and see...
Keep runnin' now,
I run and see,
Keep runnin' now,
Run and see,
Runnin' now,
Run and see,
Runnin' now...}
`)
