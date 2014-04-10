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

package junit_user

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/junit_user/result"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"os"
	"path/filepath"
)

type (
	Tool struct {
		name    string
		pkg     string
		dir     string
		current *TestFile
		runner  *junit.Tool
	}
	TestFile struct {
		id, testId bson.ObjectId
		*tool.Target
	}

	Result struct {
		*junit.Result
	}
)

func (t *TestFile) Update(f *project.File) {
	t.id = f.Id
	t.Target = tool.NewTarget(f.Name, f.Package, t.Target.Dir, tool.JAVA)
}

func New(test *junit.Test, dir string) *Tool {
	return &Tool{
		name: test.Name,
		pkg:  test.Package,
		dir:  dir,
		current: &TestFile{
			id:     test.Id,
			testId: test.Id,
			Target: tool.NewTarget(test.Name, test.Package, filepath.Join(dir, test.Id.Hex()), tool.JAVA),
		},
	}
}

//Lang is Java
func (t *Tool) Lang() tool.Language {
	return tool.JAVA
}

func (t *Tool) Name() string {
	return t.current.Name
}

func (t *Tool) Run(fileId bson.ObjectId, target *tool.Target) (tool.ToolResult, error) {
	f, e := db.File(bson.M{db.ID: fileId}, bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	fs, e := db.Files(bson.M{project.TYPE: project.TEST, project.NAME: t.name, project.TIME: bson.M{db.LT: f.Time}}, bson.M{db.ID: 1}, "-"+db.TIME)
	if e != nil {
		return nil, e
	}
	if len(fs) == 0 {
		return nil, fmt.Errorf("No user implementations of %s found for %s.", t.name, f.Name)
	}
	if fs[0].Id != t.current.id {
		if e = t.update(fs[0].Id); e != nil {
			return nil, e
		}
	}
	r, e := t.runner.Run(fileId, target)
	if e != nil {
		return nil, e
	}
	return result.New(t.current.id, r)
}

func (t *Tool) update(newTest bson.ObjectId) error {
	if util.Exists(t.current.FilePath()) {
		if e := os.Remove(t.current.FilePath()); e != nil {
			return e
		}
	}
	tf, e := db.File(bson.M{db.ID: newTest}, nil)
	if e != nil {
		return e
	}
	t.current.Update(tf)
	jt := &junit.Test{
		Id:      t.current.testId,
		Name:    tf.Name,
		Package: tf.Package,
		Type:    junit.USER,
		Test:    tf.Data,
		Data:    nil,
	}
	t.runner, e = junit.New(jt, t.dir)
	return e
}
