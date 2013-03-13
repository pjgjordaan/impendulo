package db

import (
	"bytes"
	"encoding/json"
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/utils"
)

func Read(uname string) (info *client.ClientData, err error) {
	data, err := utils.ReadFile(dbName(uname))
	if err == nil {
		err = json.Unmarshal(data, &info)
	}
	return info, err

}
func Add(uname string, info *client.ClientData) (err error) {
	data, err := json.Marshal(info)
	if err == nil {
		buff := new(bytes.Buffer)
		buff.Write(data)
		err = utils.WriteFile(dbName(uname), buff)
	}
	return err
}

func dbName(uname string) string {
	return utils.DB_PATH + utils.SEP + uname + ".json"
}
