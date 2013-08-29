package jpf

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"bytes"
	"strings"
	"fmt"
)

//Config represents a JPF configuration file.
type Config struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Name      string        "name"
	User      string        "user"
	Time      int64         "time"
	Data      []byte        "data"
}

func (this *Config) TypeName() string {
	return "jpf configuration file"
}

func (this *Config) String() string {
	return "Type: project.Config; Id: " + this.Id.Hex() +
		"; ProjectId: " + this.ProjectId.Hex() +
		"; Name: " + this.Name + "; User: " + this.User +
		"; Time: " + util.Date(this.Time)
}

//NewFile
func NewConfig(projectId bson.ObjectId, name, user string, data []byte) *Config {
	id := bson.NewObjectId()
	return &Config{
		Id: id, 
		ProjectId: projectId, 
		Name: name, 
		User: user, 
		Time: util.CurMilis(), 
		Data: data,
	}
}


func JPFBytes(config map[string] []string)(ret []byte, err error){
	buff := new(bytes.Buffer)
	for key, val := range config{
		_, err = buff.WriteString(fmt.Sprintf("%s = %s\n", key, strings.Join(val, ",")))
		if err != nil{
			return
		}
	}
	ret = buff.Bytes()
	return
}