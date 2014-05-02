package make

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
)

type (
	Makefile struct {
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		Time      int64         `bson:"time"`
		Data      []byte        `bson:"data"`
	}
)

func NewMakefile(projectId bson.ObjectId, data []byte) *Makefile {
	id := bson.NewObjectId()
	return &Makefile{id, projectId, util.CurMilis(), data}
}
