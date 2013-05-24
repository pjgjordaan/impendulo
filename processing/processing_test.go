package processing

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"labix.org/v2/mgo/bson"
	"testing"
	"os"
	"path/filepath"
	"bytes"
	"reflect"
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
	dir := filepath.Join(os.TempDir(), s.Id.Hex(), SRC)
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
	orig := map[bson.ObjectId]bool{bson.NewObjectId():true, bson.NewObjectId():false, bson.NewObjectId():false, bson.NewObjectId():true}
	err := saveActive(orig)
	if err != nil{
		t.Error(err)
	}
	ret := getStored()
	if !reflect.DeepEqual(orig, ret){
		t.Error("Maps not equal")
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
