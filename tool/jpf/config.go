package jpf

import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"strings"
)

type (
	//Config represents a JPF configuration.
	//This means that it contains configured JPF properties specific to running a certain project.
	Config struct {
		Id        bson.ObjectId "_id"
		ProjectId bson.ObjectId "projectid"
		User      string        "user"
		Time      int64         "time"
		//Contains configured JPF properties
		Data []byte "data"
	}
	empty struct{}
)

var (
	//reserved are JPF properties which are not allowed to be set by the user.
	reserved = map[string]empty{
		"search.class":                  empty{},
		"listener":                      empty{},
		"target":                        empty{},
		"report.publisher":              empty{},
		"report.xml.class":              empty{},
		"report.xml.file":               empty{},
		"classpath":                     empty{},
		"report.xml.start":              empty{},
		"report.xml.transition":         empty{},
		"report.xml.constraint":         empty{},
		"report.xml.property_violation": empty{},
		"report.xml.show_steps":         empty{},
		"report.xml.show_method":        empty{},
		"report.xml.show_code":          empty{},
		"report.xml.finished":           empty{},
	}
)

//Allowed checks whether a JPF property is allowed to be configured.
func Allowed(key string) bool {
	_, ok := reserved[key]
	return !ok
}

//String
func (this *Config) String() string {
	return "Type: project.Config; Id: " + this.Id.Hex() +
		"; ProjectId: " + this.ProjectId.Hex() +
		"; User: " + this.User + "; Time: " + util.Date(this.Time)
}

//NewConfig creates a new JPF configuration for a certain project.
func NewConfig(projectId bson.ObjectId, user string, data []byte) *Config {
	id := bson.NewObjectId()
	return &Config{
		Id:        id,
		ProjectId: projectId,
		User:      user,
		Time:      util.CurMilis(),
		Data:      data,
	}
}

//JPFBytes converts a map of JPF properties to a byte array which can be used when running JPF.
func JPFBytes(config map[string][]string) (ret []byte, err error) {
	buff := new(bytes.Buffer)
	for key, val := range config {
		_, err = buff.WriteString(fmt.Sprintf("%s = %s\n", key, strings.Join(val, ",")))
		if err != nil {
			return
		}
	}
	ret = buff.Bytes()
	return
}
