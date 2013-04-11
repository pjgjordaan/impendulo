package utils

import (
	"labix.org/v2/mgo/bson"
	"net/rpc"
)

const PROTOCOL = "tcp"
const ADDRESS = "localhost"
const PORT = "6000"

func ProcessFile(file bson.M) error {
	return Process("Server.Process", file)
}

func Process(function string, data bson.M) (err error) {
	c, err := rpc.Dial(PROTOCOL, ADDRESS+":"+PORT)
	if err == nil {
		err = c.Call(function, data, nil)
	}
	return err
}
