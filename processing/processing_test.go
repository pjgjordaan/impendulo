package processing

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"testing"
	"os"
	"path/filepath"
	"bytes"
	"reflect"
	"time"
	"math/rand"
)

func TestAddResult(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	file := submission.NewFile(bson.NewObjectId(), map[string]interface{}{}, fileData)
	err := db.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	res := tool.NewResult(file.Id, bson.NewObjectId(), "dummy", "dummy_w", "dummy_e", fileData, fileData, nil)
	err = AddResult(res)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{submission.ID: file.Id}
	dbFile, err := db.GetFile(matcher)
	if dbFile.Results["dummy"] != res.Id {
		t.Error("File not updated")
	}
	matcher = bson.M{submission.ID: res.Id}
	dbRes, err := db.GetResult(matcher)
	if !res.Equals(dbRes) {
		t.Error("Result not added correctly")
	}
}

func TestExtractFile(t *testing.T){
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := submission.NewSubmission("Triangle", "user", submission.FILE_MODE, "java")
	err := db.AddSubmission(s)
	if err != nil{
		t.Error(err)
	}
	info := bson.M{submission.TIME: 1000, submission.TYPE: submission.SRC, submission.MOD: 'c', submission.NAME: "Triangle.java", submission.FTYPE: "java", submission.PKG: "triangle", submission.NUM: 100}
	f := submission.NewFile(s.Id, info, fileData)
	dir := filepath.Join(os.TempDir(), s.Id.Hex())
	defer os.RemoveAll(dir)
	ti, err := ExtractFile(f, dir)
	if err != nil{
		t.Error(err)
	}
	expected := tool.NewTarget("Triangle", "Triangle.java", "java", "triangle", dir)
	if !expected.Equals(ti){
		t.Error("Targets not equivalent")
	}
	file, err := os.Open(filepath.Join(dir, filepath.Join("triangle","Triangle.java")))
	if err != nil{
		t.Error(err)
	}
	buff := new(bytes.Buffer)
	_, err = buff.ReadFrom(file)
	if err != nil{
		t.Error(err)
	}
	if !bytes.Equal(fileData, buff.Bytes()){
		t.Error("Data not equivalent")
	}
}

func TestStore(t *testing.T){
	fname := "test0.gob"
	orig := genMap()
	defer os.Remove(filepath.Join(util.BaseDir(), fname))
	err := saveActive(fname, orig)
	if err != nil{
		t.Error(err)
	}
	ret := getStored(fname)
	if !reflect.DeepEqual(orig, ret){
		t.Error("Maps not equal")
	}
}


func TestMonitor(t *testing.T){
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
	for k, v := range subMap{
		busy <- k
		if !v {
			count ++
			go func(id bson.ObjectId){
				time.Sleep(time.Millisecond*time.Duration(r.Intn(100)))
				done <- id
				completed <- true
			}(k)
		}
	}
	for i := 0; i < count; i ++{
		<-completed
	}
	quit <- os.Interrupt
	//Have to wait for file to be 
	time.Sleep(time.Second)	
	retrieved := getStored(fname)
	for k,v := range subMap{
		if v != retrieved[k]{
			t.Error("Map values did not match for submission:", k, v, retrieved[k])
		}
	}
}

func genMap()map[bson.ObjectId]bool{
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idMap := make(map[bson.ObjectId]bool)
	for i := 0; i < 100; i++{
		idMap[bson.NewObjectId()] = r.Float64() > 0.5
	}
	return idMap
}

func TestProcessStored(t *testing.T){
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	sub := submission.NewSubmission("Triangle", "user", submission.FILE_MODE, "java")
	err := db.AddSubmission(sub)
	if err != nil{
		t.Error(err)
	}
	ids := make(map[bson.ObjectId]bool)
	for i := 0; i < 5; i ++{
		info := bson.M{submission.TIME: 1000+i, submission.TYPE: submission.SRC, submission.MOD: 'c', submission.NAME: "Triangle.java", submission.FTYPE: "java", submission.PKG: "triangle", submission.NUM: i}
		f := submission.NewFile(sub.Id, info, fileData)
		err := db.AddFile(f)
		if err != nil{
			t.Error(err)
		}
		ids[f.Id] = true
	}
	subChan := make(chan *submission.Submission)
	fileChan := make(chan *submission.File)
	go ProcessStored(sub.Id, subChan, fileChan)
	gotSub := false
loop : for{
		select{
		case s := <-subChan:
			if !sub.Equals(s){
				t.Error("Submissions not equal", sub, s)
			}
			if gotSub{
				break loop
			}
			gotSub = true
		case file := <- fileChan:
			if !ids[file.Id]{
				t.Error("Unknown id", file.Id)
			} else {
				f, err := db.GetFile(bson.M{submission.ID: file.Id})
				if err != nil{
					t.Error(err)
				}
				if !f.Equals(file){
					t.Error("Files not equal")
				}
			}
			delete(ids, file.Id)
		}
	}
	if len(ids) > 0{
		t.Error("All files not received", ids)
	}
}

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
