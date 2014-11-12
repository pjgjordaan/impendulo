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

package test

import (
	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"path/filepath"
)

func TestTools(w *Worker, f *project.File) ([]tool.T, error) {
	switch project.Language(w.project.Lang) {
	case project.JAVA:
		return javaTestTools(w, f)
	case project.C:
		return cTestTools(w, f), nil
	}
	//Only Java is supported so far...
	return nil, fmt.Errorf("no tools found for %s language", w.project.Lang)
}

func cTestTools(w *Worker, f *project.File) []tool.T {
	return []tool.T{}
}

func javaTestTools(w *Worker, f *project.File) ([]tool.T, error) {
	a := make([]tool.T, 0, 2)
	target := tool.NewTarget(f.Name, f.Package, filepath.Join(w.toolDir, f.Id.Hex()), project.JAVA)
	if e := util.SaveFile(target.FilePath(), f.Data); e != nil {
		return nil, e
	}
	t, e := db.JUnitTest(bson.M{db.PROJECTID: w.project.Id, db.NAME: f.Name, db.TYPE: junit.USER}, bson.M{db.TARGET: 1})
	if e != nil {
		return nil, e
	}
	ju, e := junit.New(target, t.Target, w.toolDir, f.Id)
	if e != nil {
		return nil, e
	}
	a = append(a, ju)
	ja, e := jacoco.New(w.rootDir, w.srcDir, target, t.Target, f.Id)
	if e != nil {
		return nil, e
	}
	return append(a, ja), nil
}
