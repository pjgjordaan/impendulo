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
)

func (this *TestFile) Update(file *project.File) {
	this.id = file.Id
	this.Target = tool.NewTarget(file.Name, file.Package, this.Target.Dir, tool.JAVA)
}

func New(test *junit.Test, dir string) (junit *Tool, err error) {
	junit = &Tool{
		name: test.Name,
		pkg:  test.Package,
		dir:  dir,
		current: &TestFile{
			id:     test.Id,
			testId: test.Id,
			Target: tool.NewTarget(test.Name, test.Package, filepath.Join(dir, test.Id.Hex()), tool.JAVA),
		},
	}
	return
}

//Lang is Java
func (this *Tool) Lang() tool.Language {
	return tool.JAVA
}

func (this *Tool) Name() string {
	return this.current.Name
}

func (this *Tool) Run(fileId bson.ObjectId, t *tool.Target) (res tool.ToolResult, err error) {
	file, err := db.File(bson.M{db.ID: fileId}, bson.M{db.DATA: 0})
	if err != nil {
		return
	}
	matcher := bson.M{project.TYPE: project.TEST, project.NAME: this.name, project.TIME: bson.M{db.LT: file.Time}}
	files, err := db.Files(matcher, bson.M{db.ID: 1}, "-"+db.TIME)
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = fmt.Errorf("No user implementations of %s found for %s.", this.name, file.Name)
		return
	}
	if files[0].Id != this.current.id {
		err = this.update(files[0].Id)
		if err != nil {
			return
		}
	}
	res, err = this.runner.Run(fileId, t)
	return
}

func (this *Tool) update(newTest bson.ObjectId) (err error) {
	if util.Exists(this.current.FilePath()) {
		err = os.Remove(this.current.FilePath())
		if err != nil {
			return
		}
	}
	testFile, err := db.File(bson.M{db.ID: newTest}, nil)
	if err != nil {
		return
	}
	this.current.Update(testFile)
	test := &junit.Test{
		Id:      this.current.testId,
		Name:    testFile.Name,
		Package: testFile.Package,
		Type:    junit.USER,
		Test:    testFile.Data,
		Data:    nil,
	}
	this.runner, err = junit.New(test, this.dir)
	return
}
