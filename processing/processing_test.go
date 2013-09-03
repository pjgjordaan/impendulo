package processing

import (
	"bytes"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func init() {
	err := config.LoadConfigs("../config.txt")
	if err != nil {
		panic(err)
	}
}

func TestExtractFile(t *testing.T) {
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
	analyser := &Analyser{proc: proc, file: file}
	err = analyser.buildTarget()
	if err != nil {
		t.Error(err)
	}
	expected := tool.NewTarget("Triangle.java", "java", "triangle", proc.srcDir)
	if !reflect.DeepEqual(expected, analyser.target) {
		t.Error("Targets not equivalent")
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

func TestEval(t *testing.T) {
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
	proc.tests, err = SetupTests(proc.sub.ProjectId, proc.toolDir)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(proc.rootDir)
	analyser := &Analyser{proc: proc, file: file}
	err = analyser.Eval()
	if err != nil {
		t.Error(err)
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
	p := project.NewProject("Test", "user", "java", []byte{})
	err = db.AddProject(p)
	if err != nil {
		t.Error(err)
	}
	n := 2
	subs := make([]*project.Submission, n)
	archives := make([]*project.File, n)
	for i, _ := range subs {
		sub := project.NewSubmission(p.Id, "user", project.ARCHIVE_MODE, util.CurMilis())
		archive := project.NewArchive(sub.Id, zipped)
		err = db.AddSubmission(sub)
		if err != nil {
			t.Error(err)
		}
		err = db.AddFile(archive)
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
	p := project.NewProject("Triangle", "user", "java", []byte{})
	err = db.AddProject(p)
	if err != nil {
		return
	}
	s := project.NewSubmission(p.Id, p.User, project.FILE_MODE, 1000)
	err = db.AddSubmission(s)
	if err != nil {
		return
	}
	file, err = project.NewFile(s.Id, fileInfo, fileData)
	if err != nil {
		return
	}
	err = db.AddFile(file)
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
