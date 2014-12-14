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
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"

	"reflect"
	"testing"
)

func TestResultGridFS(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, e := Session()
	if e != nil {
		t.Error(e)
	}
	defer s.Close()
	u := user.New("user", "password")
	if e = Add(USERS, u); e != nil {
		t.Error(e)
	}
	p := project.New("project", u.Name, "none", project.JAVA)
	if e = Add(PROJECTS, p); e != nil {
		t.Error(e)
	}
	sk := project.NewSkeleton(p.Id, "skeleton", []byte{})
	if e = Add(SKELETONS, sk); e != nil {
		t.Error(e)
	}
	a := project.NewAssignment(p.Id, sk.Id, "assignment", u.Name, 1000, 100000)
	if e = Add(ASSIGNMENTS, a); e != nil {
		t.Error(e)
	}
	sub := project.NewSubmission(p.Id, a.Id, u.Name, project.FILE_MODE, 10000)
	if e = Add(SUBMISSIONS, sub); e != nil {
		t.Error(e)
	}
	f, e := project.NewFile(sub.Id, fileInfo, fileData)
	if e != nil {
		t.Error(e)
	}
	if e = Add(FILES, f); e != nil {
		t.Error(e)
	}
	results := []result.Tooler{
		checkstyleResult(f.Id, true),
		findbugsResult(f.Id, true),
		javacResult(f.Id, true),
	}
	for _, r := range results {
		if e = AddResult(r, r.GetName()); e != nil {
			t.Error(e)
		}
		v, e := Tooler(bson.M{"_id": r.GetId()}, nil)
		if e != nil {
			t.Error(e)
		}
		r.SetReport(report(r.GetName(), r.GetId()))
		if !reflect.DeepEqual(r, v) {
			t.Error("Results not equivalent", v)
		}
	}
}

func TestGridFS(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, e := Session()
	if e != nil {
		t.Error(e)
	}
	defer s.Close()
	id := bson.NewObjectId()
	if e = AddGridFile(id, grandPrix); e != nil {
		t.Error(e)
	}
	var d []byte
	if e = GridFile(id, &d); e != nil {
		t.Error(e)
	}
	if !bytes.Equal(grandPrix, d) {
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
