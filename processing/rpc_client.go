package processing

import (
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
	"net/rpc"
	"strconv"
)

var (
	rpcAddress string
)

func AddFile(file *project.File) (err error) {
	//We only need to process source files  and archives.
	if !file.CanProcess() {
		return
	}
	client, err := rpc.DialHTTP("tcp", rpcAddress)
	if err != nil {
		return
	}
	req := &FileRequest{
		Id:    file.Id,
		SubId: file.SubId,
	}
	err = client.Call("RPCHandler.AddFile", req, &Empty{})
	return
}

func StartSubmission(subId bson.ObjectId) (err error) {
	return toggleSubmission(subId, true)
}

func EndSubmission(subId bson.ObjectId) error {
	return toggleSubmission(subId, false)
}

func toggleSubmission(subId bson.ObjectId, start bool) (err error) {
	client, err := rpc.DialHTTP("tcp", rpcAddress)
	if err != nil {
		return
	}
	req := &SubRequest{
		Id:    subId,
		Start: start,
	}
	err = client.Call("RPCHandler.ToggleSubmission", req, &Empty{})
	return
}

func GetStatus() (ret *Status, err error) {
	client, err := rpc.DialHTTP("tcp", rpcAddress)
	if err != nil {
		return
	}
	ret = new(Status)
	err = client.Call("RPCHandler.GetStatus", Empty{}, ret)
	return
}

func SetClientAddress(address string, port int) {
	rpcAddress = address + ":" + strconv.Itoa(port)
}
