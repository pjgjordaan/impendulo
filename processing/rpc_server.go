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
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type (
	RPCHandler struct {
		sub    chan *SubRequest
		file   chan *FileRequest
		status chan Status
	}
)

func setupRPC(port int, subCh chan *SubRequest, fileCh chan *FileRequest, statusCh chan Status) (err error) {
	handler := &RPCHandler{
		sub:    subCh,
		file:   fileCh,
		status: statusCh,
	}
	rpc.Register(handler)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err == nil {
		go http.Serve(l, nil)
	}
	return
}

//GetStatus retrieves Impendulo's current processing status.
func (this *RPCHandler) GetStatus(nothing Empty, ret *Status) error {
	this.status <- Status{0, 0}
	s := <-this.status
	ret.Files, ret.Submissions = s.Files, s.Submissions
	return nil
}

//AddFile sends a file id to be processed.
func (this *RPCHandler) AddFile(req *FileRequest, nothing *Empty) error {
	req.Response = make(chan error)
	//We send the file's db id as well as the id of the submission to which it belongs.
	this.file <- req
	//Return any errors which occured while adding the file.
	return <-req.Response
}

//ToggleSubmission signals toggles the current submission state.
func (this *RPCHandler) ToggleSubmission(req *SubRequest, nothing *Empty) error {
	req.Response = make(chan error)
	this.sub <- req
	return <-req.Response
}

func (this *RPCHandler) RedoSubmission(subId bson.ObjectId, nothing *Empty) error {
	go redoSubmission(subId, this.sub, this.file)
	return nil
}

func redoSubmission(subId bson.ObjectId, subChan chan *SubRequest, fileChan chan *FileRequest) {
	sreq := &SubRequest{
		Id:       subId,
		Response: make(chan error),
		Start:    true,
	}
	subChan <- sreq
	err := <-sreq.Response
	if err != nil {
		util.Log(err)
	}
	matcher := bson.M{db.SUBID: subId}
	selector := bson.M{db.DATA: 0}
	files, err := db.Files(matcher, selector, db.TIME)
	if err != nil {
		util.Log(err)
	}
	for _, f := range files {
		freq := &FileRequest{
			Id:       f.Id,
			SubId:    subId,
			Response: make(chan error),
		}
		fileChan <- freq
		err = <-freq.Response
		if err != nil {
			util.Log(err)
		}
	}
	sreq = &SubRequest{
		Id:       subId,
		Response: make(chan error),
		Start:    false,
	}
	subChan <- sreq
	err = <-sreq.Response
	if err != nil {
		util.Log(err)
	}
}
