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

package processing

import (
	"bytes"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestProcessFile(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	file, err := setupFile()
	if err != nil {
		t.Error(err)
	}
	proc, err := NewProcessor(file.SubId)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(proc.rootDir)
	err = proc.ProcessFile(file)
	if err != nil {
		t.Error(err)
	}
	stored, err := os.Open(filepath.Join(proc.srcDir,
		filepath.Join("triangle", "Triangle.java")))
	if err != nil {
		t.Error(err)
	}
	buff := new(bytes.Buffer)
	_, err = buff.ReadFrom(stored)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(fileData, buff.Bytes()) {
		t.Error("Data not equivalent")
	}
}

func TestArchive(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	name := "_za.ac.sun.ac.za.Triangle_src_triangle_Triangle.java_"
	time := 1256033823717
	num := 8583
	toZip := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		t := strconv.Itoa(time + i*100)
		n := strconv.Itoa(num + i)
		toZip[name+t+"_"+n+"_c"] = fileData
	}
	zipped, err := util.ZipMap(toZip)
	if err != nil {
		t.Error(err)
	}
	p := project.NewProject("Test", "user", tool.JAVA, []byte{})
	err = db.Add(db.PROJECTS, p)
	if err != nil {
		t.Error(err)
	}
	n := 2
	subs := make([]*project.Submission, n)
	archives := make([]*project.File, n)
	for i, _ := range subs {
		sub := project.NewSubmission(p.Id, "user", project.ARCHIVE_MODE, util.CurMilis())
		archive := project.NewArchive(sub.Id, zipped)
		err = db.Add(db.SUBMISSIONS, sub)
		if err != nil {
			t.Error(err)
		}
		err = db.Add(db.FILES, archive)
		if err != nil {
			t.Error(err)
		}
		subs[i] = sub
		archives[i] = archive
	}
	go func() {
		for j, sub := range subs {
			StartSubmission(sub.Id)
			AddFile(archives[j])
			EndSubmission(sub.Id)
		}
		Shutdown()
	}()
	Serve(10)
	return
}

func setupFile() (file *project.File, err error) {
	p := project.NewProject("Triangle", "user", tool.JAVA, []byte{})
	err = db.Add(db.PROJECTS, p)
	if err != nil {
		return
	}
	s := project.NewSubmission(p.Id, p.User, project.FILE_MODE, 1000)
	err = db.Add(db.SUBMISSIONS, s)
	if err != nil {
		return
	}
	file, err = project.NewFile(s.Id, fileInfo, fileData)
	if err != nil {
		return
	}
	err = db.Add(db.FILES, file)
	return
}

func genMap() map[bson.ObjectId]bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idMap := make(map[bson.ObjectId]bool)
	for i := 0; i < 100; i++ {
		idMap[bson.NewObjectId()] = r.Float64() > 0.5
	}
	return idMap
}

var fileInfo = bson.M{
	project.TIME: 1000,
	project.TYPE: project.SRC,
	project.NAME: "Triangle.java",
	project.PKG:  "triangle",
}

var fileData = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)
