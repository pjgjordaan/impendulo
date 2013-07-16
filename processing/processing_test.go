package processing

import (
	"bytes"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/javac"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
	"fmt"
)

func TestAddResult(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	file, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = db.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	res := javac.NewResult(file.Id, fileData)
	err = AddResult(res)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{project.ID: file.Id}
	dbFile, err := db.GetFile(matcher, nil)
	if dbFile.Results[javac.NAME] != res.Id {
		t.Error("File not updated")
	}
	matcher = bson.M{project.ID: res.Id}
	dbRes, err := db.GetJavacResult(matcher, nil)
	if !res.Equals(dbRes) {
		t.Error("Result not added correctly")
	}
	fmt.Println("TestAddResult Success")
}

func TestExtractFile(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	p := project.NewProject("Triangle", "user", "java")
	err := db.AddProject(p)
	if err != nil {
		t.Error(err)
	}
	s := project.NewSubmission(p.Id, p.User, project.FILE_MODE, 1000)
	err = db.AddSubmission(s)
	if err != nil {
		t.Error(err)
	}
	file, err := project.NewFile(s.Id, fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	proc := NewProcessor(s, make(chan bson.ObjectId))
	defer os.RemoveAll(proc.rootDir)
	analyser := &Analyser{proc: proc, file: file}
	err = analyser.buildTarget()
	if err != nil {
		t.Error(err)
	}
	expected := tool.NewTarget("Triangle.java", "java", "triangle", proc.srcDir)
	if !expected.Equals(analyser.target) {
		t.Error("Targets not equivalent")
	}
	stored, err := os.Open(filepath.Join(proc.srcDir, filepath.Join("triangle", "Triangle.java")))
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
	fmt.Println("TestExtractFile Success")
}

/*func TestStore(t *testing.T) {
	fname := "test0.gob"
	orig := genMap()
	defer os.Remove(filepath.Join(util.BaseDir(), fname))
	err := saveActive(fname, orig)
	if err != nil {
		t.Error(err)
	}
	ret := getStored(fname)
	if !reflect.DeepEqual(orig, ret) {
		t.Error("Maps not equal")
	}
}

func TestMonitor(t *testing.T) {
	fname := "test1.gob"
	busy := make(chan bson.ObjectId)
	done := make(chan bson.ObjectId)
	quit := make(chan os.Signal)
	completed := make(chan bool)
	subMap := genMap()
	defer os.Remove(filepath.Join(util.BaseDir(), fname))
	go Monitor(fname, busy, done, quit)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 0
	for k, v := range subMap {
		busy <- k
		if !v {
			count++
			go func(id bson.ObjectId) {
				time.Sleep(time.Millisecond * time.Duration(r.Intn(100)))
				done <- id
				completed <- true
			}(k)
		}
	}
	for i := 0; i < count; i++ {
		<-completed
	}
	quit <- os.Interrupt
	//Have to wait for file to be
	time.Sleep(time.Second)
	retrieved := getStored(fname)
	for k, v := range subMap {
		if v != retrieved[k] {
			t.Error("Map values did not match for submission:", k, v, retrieved[k])
		}
	}
}*/

func genMap() map[bson.ObjectId]bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idMap := make(map[bson.ObjectId]bool)
	for i := 0; i < 100; i++ {
		idMap[bson.NewObjectId()] = r.Float64() > 0.5
	}
	return idMap
}

func TestProcessStored(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	p := project.NewProject("Triangle", "user", "java")
	err := db.AddProject(p)
	if err != nil {
		t.Error(err)
	}
	sub := project.NewSubmission(p.Id, p.User, project.FILE_MODE, 1000)
	err = db.AddSubmission(sub)
	if err != nil {
		t.Error(err)
	}
	files := make(map[bson.ObjectId]*project.File)
	for i := 0; i < 5; i++ {
		info := bson.M{project.TIME: 1000 + i, project.TYPE: project.SRC, project.MOD: 'c', project.NAME: "Triangle.java", project.FTYPE: "java", project.PKG: "triangle", project.NUM: i}
		file, err := project.NewFile(sub.Id, info, fileData)
		if err != nil {
			t.Error(err)
		}
		err = db.AddFile(file)
		if err != nil {
			t.Error(err)
		}
		file, err = db.GetFile(bson.M{project.ID: file.Id}, nil)
		if err != nil {
			t.Error(err)
		}
		files[file.Id] = file
	}
	go ProcessStored(sub.Id)
	gotSub := false
loop:
	for {
		select {
		case s := <-subChan:
			if !sub.Equals(s) {
				t.Error("Submissions not equal", sub, s)
			}
			if gotSub {
				break loop
			}
			gotSub = true
		case fileId := <-fileChan:
			if sent, ok := files[fileId.id]; ok{ 
				stored, err := db.GetFile(bson.M{project.ID: fileId.id}, nil)
				if err != nil {
					t.Error(err)
				}
				if !sent.Equals(stored) {
					t.Error("Files not equal")
				}
			}else{
				t.Error("Unknown id", fileId.id)
			} 
			delete(files, fileId.id)
		}
	}
	if len(files) > 0 {
		t.Error("All files not received", files)
	}
	fmt.Println("TestProcessStored Success")
}

var fileInfo = bson.M{project.TIME: 1000, project.TYPE: project.SRC, project.MOD: "c", project.NAME: "Triangle.java", project.FTYPE: "java", project.PKG: "triangle", project.NUM: 1000}

var fileData = []byte(`Ahm Knêma
Hörr Néhêm
Ak Ëhntëhtt

Hïm Ëk Oğërr-Ré
Ïk Ün Zaahm
Ëkëm M'd/ëm Zëhënn

/ë ïlï /ë /ë

Wi wëhlö soworï
Wi wëhlö soworï

Rind/ë, rind/ë

Ô dëhndo dö wilëhndi
Doweri ï dëhndö
Döh wëh lëhndoï
Dëh wilëhn dëh weloï
Dowë lëh lëhndï
Dowëleh
Rind/ë dï ir
Rind/ë wëh loï
Rind/ë Rind/ë Rind/ë Rind/ë

Ô dëhndi dö wilëhndoï
Dowëlëh ô dëhndi
Dowëh loï
Dëh wilëhndi i
Doweri sündo
Dowelëh wëhloï
Dowëri wëri wëh wëhloï
Dö wëh wëri wï lëhndï
Dowëh roï
Dowï sëh dëh lëh leöss
Dowëh rëhndï dö wo wï
Ëpraah dö li lëhnd/ë dö li lëhnd/ë
Dëh loï
Öwi sëh wi lëh ïoss
Dowëh rïn dï dö wo wï
Ëpraah dö li lëhnd/ë dö li lëhnd/ë
..Improvisation..

Rind/ë rind/ë wëloï
Ëhn dï ïowëh ïošaa
Ëhn dëhndï loï
Siwehn dëhn

Loï wëhlö soworï
Loï wëhlö soworï
Lëhnsoï, dëh wö soworïn döwï
Ö wëh rïn dö sündo loï dëh
Rï dëhn dowëh roï `)
