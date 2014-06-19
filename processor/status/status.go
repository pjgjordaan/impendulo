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
		Submissions map[string]map[string]util.E
	}
)

func New() *S {
	return &S{FileCount: 0, Submissions: make(map[string]map[string]util.E)}
}

func (s *S) Update(r *request.R) error {
	switch r.Type {
	case request.FILE_ADD:
		return s.addFile(r)
	case request.FILE_REMOVE:
		return s.removeFile(r)
	case request.SUBMISSION_START:
		return s.addSubmission(r)
	case request.SUBMISSION_STOP:
		return s.removeSubmission(r)
	default:
		return fmt.Errorf("unknown request type %d", r.Type)
	}
}

func (s *S) removeSubmission(r *request.R) error {
	sk := r.SubId.Hex()
	if fm, ok := s.Submissions[sk]; !ok {
		return fmt.Errorf("submission %s does not exist", sk)
	} else if len(fm) > 0 {
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
	s.Submissions[sk] = make(map[string]util.E)
	return nil
}

func (s *S) addFile(r *request.R) error {
	sk, fk := r.SubId.Hex(), r.FileId.Hex()
	fm, ok := s.Submissions[sk]
	if !ok {
		return fmt.Errorf("submission %s does not exist for file %s", sk, fk)
	}
	if _, ok = fm[fk]; ok {
		return fmt.Errorf("file %s exists for submission %s", fk, sk)
	}
	fm[fk] = util.E{}
	s.FileCount++
	return nil
}

func (s *S) removeFile(r *request.R) error {
	sk, fk := r.SubId.Hex(), r.FileId.Hex()
	fm, ok := s.Submissions[sk]
	if !ok {
		return fmt.Errorf("submission %s does not exist for file %s", sk, fk)
	}
	if _, ok = fm[fk]; !ok {
		return fmt.Errorf("file %s does not exist for submission %s", fk, sk)
	}
	delete(fm, fk)
	s.FileCount--
	return nil
}

func (s *S) Idle() bool {
	return len(s.Submissions) == 0
}
