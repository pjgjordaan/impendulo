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

//Package processing provides functionality for running a submission and its snapshots
//through the Impendulo tool suite.
package request

import (
	"fmt"

	"labix.org/v2/mgo/bson"
)

type (

	//Request is used to carry requests to process submissions and files.
	R struct {
		SubId, FileId bson.ObjectId
		Type          Type
	}
	Type uint8
)

const (
	SUBMISSION_START Type = iota
	SUBMISSION_STOP
	FILE_ADD
	FILE_REMOVE
)

func (t Type) String() string {
	switch t {
	case SUBMISSION_START:
		return "SUBMISSION_START REQUEST"
	case SUBMISSION_STOP:
		return "SUBMISSION_STOP REQUEST"
	case FILE_ADD:
		return "FILE_ADD REQUEST"
	case FILE_REMOVE:
		return "FILE_REMOVE REQUEST"
	default:
		return fmt.Sprintf("UNKNOWN REQUEST %d", t)
	}
}

func (r *R) Valid() error {
	switch r.Type {
	case SUBMISSION_START, SUBMISSION_STOP, FILE_ADD, FILE_REMOVE:
		if !bson.IsObjectIdHex(r.SubId.Hex()) {
			return fmt.Errorf("Request Submission ID %s is not a valid ObjectId", r.SubId.Hex())
		} else if !bson.IsObjectIdHex(r.FileId.Hex()) {
			return fmt.Errorf("Request File ID %s is not a valid ObjectId", r.FileId.Hex())
		}
		return nil
	default:
		return fmt.Errorf("unknown Request Type %d", r.Type)
	}
}

func StopSubmission(sid bson.ObjectId) *R {
	return &R{SubId: sid, FileId: sid, Type: SUBMISSION_STOP}
}

func StartSubmission(sid bson.ObjectId) *R {
	return &R{SubId: sid, FileId: sid, Type: SUBMISSION_START}
}

func AddFile(fid, sid bson.ObjectId) *R {
	return &R{SubId: sid, FileId: fid, Type: FILE_ADD}
}

func RemoveFile(fid, sid bson.ObjectId) *R {
	return &R{SubId: sid, FileId: fid, Type: FILE_REMOVE}
}
