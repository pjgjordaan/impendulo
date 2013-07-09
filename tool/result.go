package tool

import (
	"html/template"
	"labix.org/v2/mgo/bson"
)

//Result describes a tool or test's results for a given file.
type Result interface {
	GetHTML() template.HTML
	Success() bool
	Name() string
	GetId() bson.ObjectId
	GetFileId() bson.ObjectId
	String() string
}
