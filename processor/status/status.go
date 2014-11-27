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

package status

import (
	"fmt"

	"github.com/godfried/impendulo/processor/request"
	"github.com/godfried/impendulo/util"
)

type (

	//S is used to indicate a change in the files or
	//submissions being processed. It is also used to retrieve the current
	//number of files and submissions being processed.
	S struct {
		FileCount   int
		Submissions map[string]*SubInfo
	}
	SubInfo struct {
		Src     util.Set
		Test    util.Set
		Archive util.Set
	}
)

func NewSubInfo() *SubInfo {
	return &SubInfo{Src: util.NewSet(), Test: util.NewSet(), Archive: util.NewSet()}
}

func (s *SubInfo) FileCount() int {
	return len(s.Src) + len(s.Test) + len(s.Archive)
}

func (s *SubInfo) Empty() bool {
	return len(s.Src) == 0 && len(s.Test) == 0 && len(s.Archive) == 0
}

func (s *SubInfo) Add(r *request.R) error {
	switch r.Type {
	case request.SRC_ADD:
		s.Src.Add(r.FileId.Hex())
	case request.TEST_ADD:
		s.Test.Add(r.FileId.Hex())
	case request.ARCHIVE_ADD:
		s.Archive.Add(r.FileId.Hex())
	default:
		return fmt.Errorf("unsupported type %s", r.Type)
	}
	return nil
}

func (s *SubInfo) Remove(r *request.R) error {
	switch r.Type {
	case request.SRC_REMOVE:
		delete(s.Src, r.FileId.Hex())
	case request.TEST_REMOVE:
		delete(s.Test, r.FileId.Hex())
	case request.ARCHIVE_REMOVE:
		delete(s.Archive, r.FileId.Hex())
	default:
		return fmt.Errorf("unsupported type %s", r.Type)
	}
	return nil
}

func New() *S {
	return &S{FileCount: 0, Submissions: make(map[string]*SubInfo)}
}

func (s *S) Update(r *request.R) error {
	switch r.Type {
	case request.SRC_ADD, request.TEST_ADD, request.ARCHIVE_ADD:
		return s.addFile(r)
	case request.SRC_REMOVE, request.TEST_REMOVE, request.ARCHIVE_REMOVE:
		return s.removeFile(r)
	case request.SUBMISSION_START:
		return s.addSubmission(r)
	case request.SUBMISSION_STOP:
		return s.removeSubmission(r)
	default:
		return fmt.Errorf("unknown request type %s", r.Type)
	}
}

func (s *S) removeSubmission(r *request.R) error {
	sk := r.SubId.Hex()
	if si, ok := s.Submissions[sk]; !ok {
		return fmt.Errorf("submission %s does not exist", sk)
	} else if !si.Empty() {
		return fmt.Errorf("submission %s still has active files", sk)
	}
	delete(s.Submissions, sk)
	return nil
}

func (s *S) addSubmission(r *request.R) error {
	sk := r.SubId.Hex()
	if _, ok := s.Submissions[sk]; ok {
		return fmt.Errorf("submission %s already exists", sk)
	}
	s.Submissions[sk] = NewSubInfo()
	return nil
}

func (s *S) addFile(r *request.R) error {
	sk := r.SubId.Hex()
	if si, ok := s.Submissions[sk]; !ok {
		return fmt.Errorf("submission %s does not exist", sk)
	} else if e := si.Add(r); e != nil {
		return e
	}
	s.FileCount++
	return nil
}

func (s *S) removeFile(r *request.R) error {
	sk := r.SubId.Hex()
	if si, ok := s.Submissions[sk]; !ok {
		return fmt.Errorf("submission %s does not exist", sk)
	} else if e := si.Remove(r); e != nil {
		return e
	}
	s.FileCount--
	return nil
}

func (s *S) Idle() bool {
	return len(s.Submissions) == 0
}
