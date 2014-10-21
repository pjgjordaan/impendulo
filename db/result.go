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

package db

import (
	"fmt"

	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/tool/result"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type (
	TypeHolder struct {
		Type string `bson:"type"`
	}
)

var (
	NoResultsError = fmt.Errorf("no results found")
)

//CheckstyleResult retrieves a Result matching
//the given interface from the active database.
func CheckstyleResult(m, sl bson.M) (*checkstyle.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *checkstyle.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e = GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

//PMDResult retrieves a Result matching
//the given interface from the active database.
func PMDResult(m, sl bson.M) (*pmd.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *pmd.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e = GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

//FindbugsResult retrieves a Result matching
//the given interface from the active database.
func FindbugsResult(m, sl bson.M) (*findbugs.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *findbugs.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e = GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

//JPFResult retrieves a Result matching
//the given interface from the active database.
func JPFResult(m, sl bson.M) (*jpf.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *jpf.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e := GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

//JUnitResult retrieves aResult matching
//the given interface from the active database.
func JUnitResult(m, sl bson.M) (*junit.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *junit.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e := GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

func JacocoResult(m, sl bson.M) (*jacoco.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *jacoco.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e := GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

//JavacResult retrieves a JavacResult matching
//the given interface from the active database.
func JavacResult(m, sl bson.M) (*javac.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *javac.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e := GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

func GCCResult(m, sl bson.M) (*gcc.Result, error) {
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var r *gcc.Result
	if e = s.DB("").C(RESULTS).Find(m).Select(sl).One(&r); e != nil {
		return nil, &GetError{"result", e, m}
	} else if !HasGridFile(r, sl) {
		return r, nil
	}
	if e := GridFile(r.GetId(), &r.Report); e != nil {
		return nil, e
	}
	return r, nil
}

func resultType(m bson.M) (string, error) {
	s, e := Session()
	if e != nil {
		return "", e
	}
	defer s.Close()
	var h *TypeHolder
	if e = s.DB("").C(RESULTS).Find(m).One(&h); e != nil {
		return "", &GetError{"result type", e, m}
	}
	return h.Type, nil
}

func Tooler(m, sl bson.M) (result.Tooler, error) {
	t, e := resultType(m)
	if e != nil {
		return nil, e
	}
	m[TYPE] = t
	switch t {
	case javac.NAME:
		return JavacResult(m, sl)
	case jpf.NAME:
		return JPFResult(m, sl)
	case findbugs.NAME:
		return FindbugsResult(m, sl)
	case pmd.NAME:
		return PMDResult(m, sl)
	case checkstyle.NAME:
		return CheckstyleResult(m, sl)
	case gcc.NAME:
		return GCCResult(m, sl)
	case jacoco.NAME:
		return JacocoResult(m, sl)
	case junit.NAME:
		return JUnitResult(m, sl)
	default:
		return nil, fmt.Errorf("unsupported result type %s", t)
	}
}

func Displayer(m, sl bson.M) (result.Displayer, error) {
	t, e := resultType(m)
	if e != nil {
		return nil, e
	}
	m[TYPE] = t
	switch t {
	case javac.NAME:
		return JavacResult(m, sl)
	case jpf.NAME:
		return JPFResult(m, sl)
	case findbugs.NAME:
		return FindbugsResult(m, sl)
	case pmd.NAME:
		return PMDResult(m, sl)
	case checkstyle.NAME:
		return CheckstyleResult(m, sl)
	case gcc.NAME:
		return GCCResult(m, sl)
	case jacoco.NAME:
		return JacocoResult(m, sl)
	case junit.NAME:
		return JUnitResult(m, sl)
	default:
		return nil, fmt.Errorf("unsupported result type %s", t)
	}
}

func Charter(m, sl bson.M) (result.Charter, error) {
	t, e := resultType(m)
	if e != nil {
		return nil, e
	}
	m[TYPE] = t
	switch t {
	case javac.NAME:
		return JavacResult(m, sl)
	case jpf.NAME:
		return JPFResult(m, sl)
	case findbugs.NAME:
		return FindbugsResult(m, sl)
	case pmd.NAME:
		return PMDResult(m, sl)
	case checkstyle.NAME:
		return CheckstyleResult(m, sl)
	case gcc.NAME:
		return GCCResult(m, sl)
	case junit.NAME:
		return JUnitResult(m, sl)
	case jacoco.NAME:
		return JacocoResult(m, sl)
	default:
		return nil, fmt.Errorf("Unsupported result type %s.", t)
	}
}

func Coder(m, sl bson.M) (result.Coder, error) {
	t, e := resultType(m)
	if e != nil {
		return nil, e
	}
	m[TYPE] = t
	switch t {
	case jpf.NAME:
		return JPFResult(m, sl)
	case findbugs.NAME:
		return FindbugsResult(m, sl)
	case pmd.NAME:
		return PMDResult(m, sl)
	case checkstyle.NAME:
		return CheckstyleResult(m, sl)
	default:
		return nil, fmt.Errorf("Unsupported result type %s.", t)
	}
}

//AddResult adds a new result to the active database.
func AddResult(r result.Tooler, n string) error {
	if r == nil {
		return fmt.Errorf("can't add nil result")
	}
	if e := AddFileResult(r.GetFileId(), n, r.GetId()); e != nil {
		return e
	}
	if r.OnGridFS() {
		if e := AddGridFile(r.GetId(), r.Reporter()); e != nil {
			return e
		}
		r.SetReport(nil)
	}
	s, e := Session()
	if e != nil {
		return e
	}
	defer s.Close()
	if e = s.DB("").C(RESULTS).Insert(r); e != nil {
		return &AddError{r.GetName(), e}
	}
	return nil
}

//AddFileResult adds or updates a result in a file's results.
func AddFileResult(fid bson.ObjectId, n string, v interface{}) error {
	return Update(FILES, bson.M{ID: fid}, bson.M{SET: bson.M{RESULTS + "." + n: v}})
}

func Charters(fid bson.ObjectId) ([]result.Charter, error) {
	f, e := File(bson.M{ID: fid}, bson.M{DATA: 0})
	if e != nil {
		return nil, e
	}
	rs := make([]result.Charter, 0, len(f.Results))
	for _, id := range f.Results {
		if _, ok := id.(bson.ObjectId); !ok {
			continue
		}
		r, e := Charter(bson.M{ID: id}, nil)
		if e != nil {
			continue
		}
		rs = append(rs, r)
	}
	return rs, nil
}

//AllResultNames retrieves all result names for a given project.
func ResultNames(sid bson.ObjectId, fname string) (map[string]map[string][]string, error) {
	mr := &mgo.MapReduce{
		Map: `function() {
	var emitted = {};
	for (r in this.results) {
            if(r in emitted || r === undefined){
               continue;		
            }		
            var sa = r.split('-');
	    var sb = sa[0].split(':');
            var t = sb[0];
            var n = sb.length >= 2 ? sb[1] : '';
            var id = sa.length >= 2 ? sa[1] : '';                 
            if(t === ''){
                continue;
            }
            emitted[r] = true;
            o = {
                type: t,
                name: n,
                fileid: id,
            };
            emit("", o);
	}
	
};`,
		Reduce: `function(name, values) {
        var r = {};
        var added = {};
        for(i in values){
            var t = values[i].type;
            var n = values[i].name;
            var id = values[i].fileid;
            if(t === undefined || t === ''){ 
                continue;
            }
            if(!(t in r)){
                r[t] = {};
            }
            if(n === undefined || n === ''){ 
                continue;
            }
            if(!(n in r)){
                r[t][n] = [];
            }
            if(id === undefined || id === ''){ 
                continue;
            }
            r[t][n].push(id);
       }
       return r;
};`,
	}
	s, e := Session()
	if e != nil {
		return nil, e
	}
	defer s.Close()
	var ns []*struct {
		Value map[string]map[string][]string `bson:"value"`
	}
	if _, e := s.DB("").C(FILES).Find(bson.M{SUBID: sid, NAME: fname}).
		Select(bson.M{NAME: 1, RESULTS: 1}).MapReduce(mr, &ns); e != nil {
		return nil, e
	} else if len(ns) == 0 {
		return nil, NoResultsError
	}
	m := ns[0].Value
	m[result.CODE] = map[string][]string{}
	m[diff.NAME] = map[string][]string{}
	return m, nil
}

func BasicResultNames() map[string]map[string][]string {
	return map[string]map[string][]string{
		result.CODE: map[string][]string{},
		diff.NAME:   map[string][]string{},
	}
}

func ProjectResults(pid bson.ObjectId) []string {
	rs := []string{javac.NAME, pmd.NAME, findbugs.NAME, checkstyle.NAME}
	if Contains(JPF, bson.M{PROJECTID: pid}) {
		rs = append(rs, jpf.NAME)
	}
	ts, e := JUnitTests(bson.M{PROJECTID: pid}, bson.M{NAME: 1})
	if e != nil {
		return rs
	}
	for _, t := range ts {
		n, _ := util.Extension(t.Name)
		rs = append(rs, junit.NAME+":"+n, jacoco.NAME+":"+n)
	}
	return rs
}

func AllResults() []string {
	rs := []string{javac.NAME, pmd.NAME, findbugs.NAME, checkstyle.NAME, jpf.NAME}
	s, e := Session()
	if e != nil {
		return rs
	}
	defer s.Close()
	var ts []string
	if e := s.DB("").C(TESTS).Find(nil).Distinct(NAME, &ts); e == nil {
		for _, t := range ts {
			n, _ := util.Extension(t)
			rs = append(rs, junit.NAME+":"+n, jacoco.NAME+":"+n)
		}
	}
	return rs
}

func UserResults(u string) []string {
	rs := []string{javac.NAME, pmd.NAME, findbugs.NAME, checkstyle.NAME}
	s, e := Session()
	if e != nil {
		return rs
	}
	defer s.Close()
	var ids []bson.ObjectId
	if e := s.DB("").C(SUBMISSIONS).Find(bson.M{USER: u}).Distinct(PROJECTID, &ids); e != nil || len(ids) == 0 {
		return rs
	}
	type q struct{}
	a := map[string]q{javac.NAME: q{}, pmd.NAME: q{}, findbugs.NAME: q{}, checkstyle.NAME: q{}}
	for _, id := range ids {
		prs := ProjectResults(id)
		for _, r := range prs {
			if _, ok := a[r]; !ok {
				a[r] = q{}
				rs = append(rs, r)
			}
		}
	}
	return rs
}

func FileResultId(sid bson.ObjectId, fname, rtipe, rname string) (bson.ObjectId, error) {
	mr := &mgo.MapReduce{
		Map: `function() {
	var added = {};
        var rtipe = "` + rtipe + `";
        var rname = "` + rname + `";
	for (n in this.results) {
                if(n in added){
                    continue;
                }
                added[n] = true;
                var sa = n.split('-');
                var sb = sa[0].split(':');
                if(sa.length < 2 || sb.length < 2 || sb[0] !== rtipe || sb[1] !== rname){
                    continue;
                }
                emit(sa[1], "");
	}
};`,
		Reduce: `function(name, vals) {
       return name;
};`,
	}
	s, e := Session()
	if e != nil {
		return "", e
	}
	defer s.Close()
	var ns []*struct {
		Value string `bson:"value"`
	}
	if _, e := s.DB("").C(FILES).Find(bson.M{SUBID: sid, NAME: fname}).
		Select(bson.M{NAME: 1, RESULTS: 1}).MapReduce(mr, &ns); e != nil {
		return "", e
	}
	var f *project.File
	for _, hex := range ns {
		id, e := convert.Id(hex.Value)
		if e != nil {
			continue
		}
		cf, e := File(bson.M{ID: id}, bson.M{ID: 1, TIME: 1})
		if e != nil {
			continue
		}
		if f == nil || cf.Time > f.Time {
			f = cf
		}
	}
	if f == nil {
		return "", fmt.Errorf("no results found for %s \u2192 %s", rtipe, rname)
	}
	return f.Id, nil
}
