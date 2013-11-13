//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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

func RedoSubmission(subId bson.ObjectId) (err error) {
	client, err := rpc.DialHTTP("tcp", rpcAddress)
	if err != nil {
		return
	}
	err = client.Call("RPCHandler.RedoSubmission", subId, &Empty{})
	return
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

func WaitIdle() (err error) {
	client, err := rpc.DialHTTP("tcp", rpcAddress)
	if err != nil {
		return
	}
	err = client.Call("RPCHandler.WaitIdle", Empty{}, &Empty{})
	return
}

func SetClientAddress(address string, port int) {
	rpcAddress = address + ":" + strconv.Itoa(port)
}
