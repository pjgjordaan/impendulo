package processing

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
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
	if !reflect.DeepEqual(res, dbRes) {
		t.Error("Result not added correctly")
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
