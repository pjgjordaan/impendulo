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
		err = AddResult(res)
		if err != nil {
			t.Error(err)
		}
		matcher := bson.M{"_id": res.GetId()}
		dbRes, err := ToolResult(res.GetName(), matcher, nil)
		if err != nil {
			t.Error(err)
		}
		res.SetData(report(res.GetName(), res.GetId()))
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
