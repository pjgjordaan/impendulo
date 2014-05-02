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

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/checkstyle"
	"github.com/godfried/impendulo/tool/diff"
	"github.com/godfried/impendulo/tool/findbugs"
	"github.com/godfried/impendulo/tool/gcc"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type (
	TypeHolder struct {
		Type string `bson:"type"`
	}
)

//CheckstyleResult retrieves a Result matching
//the given interface from the active database.
func CheckstyleResult(matcher, selector bson.M) (ret *checkstyle.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = checkstyle.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

//PMDResult retrieves a Result matching
//the given interface from the active database.
func PMDResult(matcher, selector bson.M) (ret *pmd.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = pmd.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

//FindbugsResult retrieves a Result matching
//the given interface from the active database.
func FindbugsResult(matcher, selector bson.M) (ret *findbugs.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = findbugs.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

//JPFResult retrieves a Result matching
//the given interface from the active database.
func JPFResult(matcher, selector bson.M) (ret *jpf.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = jpf.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

//JUnitResult retrieves aResult matching
//the given interface from the active database.
func JUnitResult(matcher, selector bson.M) (ret *junit.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[TYPE] = junit.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func JacocoResult(matcher, selector bson.M) (ret *jacoco.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[TYPE] = jacoco.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

//JavacResult retrieves a JavacResult matching
//the given interface from the active database.
func JavacResult(matcher, selector bson.M) (ret *javac.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = javac.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func GCCResult(matcher, selector bson.M) (ret *gcc.Result, err error) {
	session, err := Session()
	if err != nil {
		return
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	matcher[NAME] = gcc.NAME
	err = c.Find(matcher).Select(selector).One(&ret)
	if err != nil {
		err = &GetError{"result", err, matcher}
	} else if HasGridFile(ret, selector) {
		err = GridFile(ret.GetId(), &ret.Report)
	}
	return
}

func resultType(matcher bson.M) (string, error) {
	session, err := Session()
	if err != nil {
		return "", err
	}
	defer session.Close()
	c := session.DB("").C(RESULTS)
	var holder *TypeHolder
	err = c.Find(matcher).One(&holder)
	if err != nil {
		return "", &GetError{"result type", err, matcher}
	}
	return holder.Type, nil
}

//ToolResult retrieves a tool.ToolResult matching
//the given interface and name from the active database.
func ToolResult(matcher, selector bson.M) (ret tool.ToolResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

//DisplayResult retrieves a tool.DisplayResult matching
//the given interface and name from the active database.
func DisplayResult(matcher, selector bson.M) (ret tool.DisplayResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

func ChartResult(matcher, selector bson.M) (ret tool.ChartResult, err error) {
	tipe, err := resultType(matcher)
	if err != nil {
		return
	}
	switch tipe {
	case javac.NAME:
		ret, err = JavacResult(matcher, selector)
	case jpf.NAME:
		ret, err = JPFResult(matcher, selector)
	case findbugs.NAME:
		ret, err = FindbugsResult(matcher, selector)
	case pmd.NAME:
		ret, err = PMDResult(matcher, selector)
	case checkstyle.NAME:
		ret, err = CheckstyleResult(matcher, selector)
	case gcc.NAME:
		ret, err = GCCResult(matcher, selector)
	case junit.NAME:
		ret, err = JUnitResult(matcher, selector)
	case jacoco.NAME:
		ret, err = JacocoResult(matcher, selector)
	default:
		err = fmt.Errorf("Unsupported result type %s.", tipe)
	}
	return
}

//AddResult adds a new result to the active database.
func AddResult(r tool.ToolResult, n string) error {
	if r == nil {
		return fmt.Errorf("Result is nil. In db/result.go.")
	}
	if e := AddFileResult(r.GetFileId(), n, r.GetId()); e != nil {
		return e
	}
	if r.OnGridFS() {
		if e := AddGridFile(r.GetId(), r.GetReport()); e != nil {
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
func AddFileResult(fileId bson.ObjectId, name string, value interface{}) error {
	return Update(FILES, bson.M{ID: fileId}, bson.M{SET: bson.M{RESULTS + "." + name: value}})
}

//ChartResults retrieves all tool.DisplayResults matching
//the given file Id from the active database.
func ChartResults(fileId bson.ObjectId) (ret []tool.ChartResult, err error) {
	file, err := File(bson.M{ID: fileId}, bson.M{DATA: 0})
	if err != nil {
		return
	}
	ret = make([]tool.ChartResult, 0, len(file.Results))
	for _, id := range file.Results {
		if _, ok := id.(bson.ObjectId); !ok {
			continue
		}
		res, err := ChartResult(bson.M{ID: id}, nil)
		if err != nil {
			err = nil
			continue
		}
		ret = append(ret, res)
	}
	return
}

//AllResultNames retrieves all result names for a given project.
func ResultNames(sid bson.ObjectId, fname string) (map[string]map[string][]interface{}, error) {
	mr := &mgo.MapReduce{
		Map: `function() {
	var results = {};
	for (n in this.results) {
		var sa = n.split('-');
		var o = {};
                var sb = sa[0].split(':');
                o.type = sb[0];
                o.name = sb.length >= 2 ? sb[1] : "";
                o.id = sa.length >= 2 ? sa[1] : "";                 
                if(!(n in results)){
			results[n] = true;
                        emit("", o);		
                }
	}
	
};`,
		Reduce: `function(n, vals) {
        var r = {};
        var added = {};
        for(i in vals){
            if(vals[i] in added){
                continue;
            }
            added[t+n] = true;
            var t = vals[i].type;
            var n = vals[i].name;
            var id = vals[i].id;
            if(!(t in r)){
                r[t] = {};
            } 
            if(n == ""){
                continue;
            }
            if(!(n in r[t])){
                r[t][n] = [];
            } 
            if(id == ""){
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
		Value map[string]map[string][]interface{} `bson:"value"`
	}
	if _, e := s.DB("").C(FILES).Find(bson.M{SUBID: sid, NAME: fname}).
		Select(bson.M{NAME: 1, RESULTS: 1}).MapReduce(mr, &ns); e != nil {
		return nil, e
	} else if len(ns) == 0 {
		return nil, fmt.Errorf("no results found")
	}
	m := ns[0].Value
	m[tool.CODE] = map[string][]interface{}{}
	m[diff.NAME] = map[string][]interface{}{}
	return m, nil
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
