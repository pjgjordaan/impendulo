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

	"github.com/godfried/impendulo/project"

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
	SRC_ADD
	TEST_ADD
	ARCHIVE_ADD
	SRC_REMOVE
	TEST_REMOVE
	ARCHIVE_REMOVE
)

func (t Type) String() string {
	switch t {
	case SUBMISSION_START:
		return "SUBMISSION_START REQUEST"
	case SUBMISSION_STOP:
		return "SUBMISSION_STOP REQUEST"
	case SRC_ADD:
		return "SRC_ADD REQUEST"
	case SRC_REMOVE:
		return "SRC_REMOVE REQUEST"
	case ARCHIVE_ADD:
		return "ARCHIVE_ADD REQUEST"
	case ARCHIVE_REMOVE:
		return "ARCHIVE_REMOVE REQUEST"
	case TEST_ADD:
		return "TEST_ADD REQUEST"
	case TEST_REMOVE:
		return "TEST_REMOVE REQUEST"
	default:
		return fmt.Sprintf("UNKNOWN REQUEST %d", t)
	}
}

func (r *R) Valid() error {
	switch r.Type {
	case SUBMISSION_START, SUBMISSION_STOP, SRC_ADD, SRC_REMOVE, ARCHIVE_ADD, ARCHIVE_REMOVE, TEST_ADD, TEST_REMOVE:
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

func AddFile(f *project.File) (*R, error) {
	switch f.Type {
	case project.SRC:
		return &R{SubId: f.SubId, FileId: f.Id, Type: SRC_ADD}, nil
	case project.TEST:
		return &R{SubId: f.SubId, FileId: f.Id, Type: TEST_ADD}, nil
	case project.ARCHIVE:
		return &R{SubId: f.SubId, FileId: f.Id, Type: ARCHIVE_ADD}, nil
	default:
		return nil, fmt.Errorf("unknown type %s", f.Type)
	}
}

func RemoveFile(f *project.File) (*R, error) {
	switch f.Type {
	case project.SRC:
		return &R{SubId: f.SubId, FileId: f.Id, Type: SRC_REMOVE}, nil
	case project.TEST:
		return &R{SubId: f.SubId, FileId: f.Id, Type: TEST_REMOVE}, nil
	case project.ARCHIVE:
		return &R{SubId: f.SubId, FileId: f.Id, Type: ARCHIVE_REMOVE}, nil
	default:
		return nil, fmt.Errorf("unknown type %s", f.Type)
	}
}
