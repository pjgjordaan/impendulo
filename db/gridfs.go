package db

import (
	"encoding/gob"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	GRIDFS_NAME = "fs"
)

//HasGridFile checks whether this query needs to get data from GridFS
func HasGridFile(result tool.ToolResult, selector bson.M) bool {
	return (selector == nil || selector[project.DATA] == 1) && result.OnGridFS()
}

//GridFile loads a GridFile matching id into a provided data structure from GridFS.
func GridFile(id, ret interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	fs := session.DB("").GridFS(GRIDFS_NAME)
	var file *mgo.GridFile
	file, err = fs.OpenId(id)
	if err != nil {
		return
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	err = dec.Decode(ret)
	return
}

//AddGridFile creates a new GridFile and stores the provided data structure in it via gob.
func AddGridFile(id, data interface{}) (err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	fs := session.DB("").GridFS(GRIDFS_NAME)
	var file *mgo.GridFile
	file, err = fs.Create("")
	if err != nil {
		return
	}
	defer file.Close()
	file.SetId(id)
	enc := gob.NewEncoder(file)
	err = enc.Encode(data)
	return
}
