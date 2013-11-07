package make

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	Makefile struct {
		Id        bson.ObjectId "_id"
		ProjectId bson.ObjectId "projectid"
		Time      int64         "time"
		Data      []byte        "data"
	}
)

func NewMakefile(projectId bson.ObjectId, data []byte) *Makefile {
	id := bson.NewObjectId()
	return &Makefile{id, projectId, util.CurMilis(), data}
}
