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

package project

import "labix.org/v2/mgo/bson"

type (
	Skeleton struct {
		Id        bson.ObjectId `bson:"_id"`
		ProjectId bson.ObjectId `bson:"projectid"`
		Name      string        `bson:"name"`
		//	Files     map[string]*SkeletonFile `bson:"files"`
		Data []byte `bson:"data"`
	}

/*	SkeletonFile struct {
		Name   string `bson:"name"`
		Ignore bool   `bson:"ignore"`
		Data   []byte `bson:"data"`
	}
	SkeletonWriter struct {
		files map[string]*SkeletonFile
	}*/
)

//NewSkeleton
func NewSkeleton(pid bson.ObjectId, n string, d []byte) *Skeleton {
	/*	w := &SkeletonWriter{make(map[string]*SkeletonFile)}
		if e := util.UnzipKV(w, d); e != nil {
			return nil, e
		}*/
	return &Skeleton{Id: bson.NewObjectId(), ProjectId: pid, Name: n, Data: d}
}

/*
func (s *Skeleton) AddIgnores(is []string) error {
	for _, i := range is {
		f, ok := s.Files[i]
		if !ok {
			return fmt.Errorf("could not locate skeleton file %s", i)
		}
		f.Ignore = true
	}
	return nil
}

func (w *SkeletonWriter) Write(k string, v []byte) error {
	w.files[k] = &SkeletonFile{Name: k, Ignore: false, Data: v}
	return nil
}

func (w *SkeletonWriter) Next() (string, []byte, error) {
	for k, v := range w.files {
		delete(w.files, k)
		return v.Name, v.Data, nil
	}
	return "", nil, fmt.Errorf("no files left")
}

func (s *Skeleton) Zip() ([]byte, error) {
	return util.ZipKV(&SkeletonWriter{s.Files})
}
*/
