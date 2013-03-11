package db

import (
	"bytes"
	"encoding/json"
	"intlola/client"
	"intlola/utils"
	"os"
)

const DB_PATH = "db"
const SEP = string(os.PathSeparator)

func Read(uname string) (*client.ClientData, error) {
	fname := dbName(uname)
	var info *client.ClientData
	data, err := utils.ReadFile(fname)
	if err == nil {
		err = json.Unmarshal(data, &info)

	}
	return info, err

}

func Add(uname string, info *client.ClientData) error {
	data, err := json.Marshal(info)
	if err == nil {
		buff := new(bytes.Buffer)
		buff.Write(data)
		err = utils.WriteFile(dbName(uname), buff)
	}
	return err
}

func dbName(uname string) string {
	return DB_PATH + SEP + uname + ".json"
}
