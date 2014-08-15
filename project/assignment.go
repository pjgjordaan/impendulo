package project

import "labix.org/v2/mgo/bson"

type (
	Assignment struct {
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		Name      string        `bson:"name"`
		User      string        `bson:"user"`
		Start     int64         `bson:"start"`
		End       int64         `bson:"end"`
	}
)

//New
func NewAssignment(projectId bson.ObjectId, name, user string, start, end int64) *Assignment {
	return &Assignment{Id: bson.NewObjectId(), ProjectId: projectId, Name: name, User: user, Start: start, End: end}
}
