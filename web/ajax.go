package web

import (
	"code.google.com/p/gorilla/pat"

	"strings"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	AJAX func(*http.Request) ([]byte, error)
)

func (a AJAX) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, e := a(r)
	if e != nil {
		util.Log(e, LOG_HANDLERS)
		b, _ = util.JSON(map[string]interface{}{"error": e.Error()})
	}
	fmt.Fprint(w, string(b))
}

func GenerateAJAX(r *pat.Router) {
	fs := map[string]AJAX{
		"chart": getChart, "usernames": getUsernames, "collections": collections,
		"skeletons": getSkeletons, "submissions": submissions,
		"langs": getLangs, "projects": getProjects, "files": ajaxFiles, "tools": ajaxTools,
		"code": ajaxCode, "users": ajaxUsers, "permissions": ajaxPerms, "comparables": ajaxComparables,
	}
	for n, f := range fs {
		r.Add("GET", "/"+n, f)
	}
}

func ajaxComparables(r *http.Request) ([]byte, error) {
	id, e := util.ReadId(r.FormValue("id"))
	if e != nil {
		return nil, e
	}
	tr, e := db.ToolResult(bson.M{db.ID: id}, nil)
	if e != nil {
		return nil, e
	}
	type c struct {
		Id, Name string
	}
	var cmp []*c
	switch tr.GetType() {
	case jacoco.NAME, junit.NAME:
		f, e := db.File(bson.M{db.ID: tr.GetFileId()}, bson.M{db.SUBID: 1})
		if e != nil {
			return nil, e
		}
		s, e := db.Submission(bson.M{db.ID: f.SubId}, bson.M{db.PROJECTID: 1})
		if e != nil {
			return nil, e
		}
		ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId}, bson.M{db.NAME: 1, db.TYPE: 1})
		if e != nil {
			return nil, e
		}
		cmp = make([]*c, 0, len(ts))
		for _, t := range ts {
			n, _ := util.Extension(t.Name)
			if n == tr.GetName() {
				continue
			}
			cmp = append(cmp, &c{tr.GetType() + ":" + n, tr.GetType() + " \u2192 " + n})
		}
	}
	return util.JSON(map[string]interface{}{"comparables": cmp})
}

func ajaxPerms(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"permissions": user.PermissionInfos()})
}

func ajaxUsers(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if n, e := GetString(r, "name"); e == nil {
		m[db.ID] = n
	}
	u, e := db.Users(m)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func ajaxCode(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if rid, e := util.ReadId(r.FormValue("resultid")); e == nil {
		tr, e := db.ToolResult(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
		if e != nil {
			return nil, e
		}
		m[db.ID] = tr.GetFileId()
	}
	if id, e := util.ReadId(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	f, e := db.File(m, bson.M{db.DATA: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
}

func collections(r *http.Request) ([]byte, error) {
	n, e := GetString(r, "db")
	if e != nil {
		return nil, e
	}
	c, e := db.Collections(n)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"collections": c})
}

func getProjects(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := util.ReadId(r.FormValue("id")); e == nil {
		m[db.ID] = pid
	}
	p, e := db.Projects(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"projects": p})
}

func ajaxTools(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	t, e := tools(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tools": t})
}

func ajaxFiles(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := util.ReadId(r.FormValue("subid")); e == nil {
		m[db.SUBID] = sid
	}
	if id, e := util.ReadId(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	format, _ := GetString(r, "format")
	var f interface{}
	var e error
	if format == "nested" {
		f, e = nestedFiles(m)
	} else {
		f, e = db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
	}
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"files": f})
}

func nestedFiles(m bson.M) (map[project.Type]map[string][]*project.File, error) {
	ts := []project.Type{project.SRC, project.LAUNCH, project.ARCHIVE, project.TEST}
	ns, e := db.FileNames(m)
	if e != nil {
		return nil, e
	}
	fm := make(map[project.Type]map[string][]*project.File)
	for _, t := range ts {
		nm := make(map[string][]*project.File)
		m[db.TYPE] = t
		for _, n := range ns {
			m[db.NAME] = n
			fs, e := db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
			if e == nil && len(fs) > 0 {
				nm[n] = fs
			}
		}
		if len(nm) > 0 {
			fm[t] = nm
		}
	}
	return fm, nil
}

func submissions(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if sid, e := util.ReadId(r.FormValue("id")); e == nil {
		m[db.ID] = sid
	}
	if pid, e := util.ReadId(r.FormValue("projectid")); e == nil {
		m[db.PROJECTID] = pid
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"submissions": s})
}

func getUsernames(r *http.Request) ([]byte, error) {
	pid, e := util.ReadId(r.FormValue("projectid"))
	var u []string
	if e != nil {
		u, e = db.Usernames(nil)
	} else {
		u, e = projectUsernames(pid)
	}
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"usernames": u})
}

func getLangs(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"langs": []tool.Language{tool.JAVA, tool.C}})
}

func getSkeletons(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	s, e := skeletons(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": s})
}

func getChart(r *http.Request) ([]byte, error) {
	e := r.ParseForm()
	if e != nil {
		return nil, e
	}
	ns, e := GetStrings(r, "file")
	if e != nil {
		return nil, e
	}
	if len(ns) != 1 {
		return nil, fmt.Errorf("invalid file names %v specified", ns)
	}
	rs, e := GetStrings(r, "result")
	if e != nil {
		return nil, e
	}
	if len(rs) != 1 {
		return nil, fmt.Errorf("invalid result names %v specified", rs)
	}
	subs, e := GetStrings(r, "submissions[]")
	if e != nil {
		return nil, e
	}
	/*	if f, e := GetStrings(r, "testfileid"); e == nil && len(f) == 1 {
			fid, e := util.ReadId(f[0])
			if e != nil {
				return nil, e
			} else {
				return srcViewChart(fid, rs[0], subs)
			}
		}
		if f, e := GetStrings(r, "srcfileid"); e == nil && len(f) == 1 {
			fid, e := util.ReadId(f[0])
			if e != nil {
				return nil, e
			} else {
				return testViewChart(fid, rs[0], ns[0], subs)
			}
		}*/
	var d ChartData
	var first bson.ObjectId
	for i, s := range subs {
		r := rs[0]
		m := bson.M{db.NAME: ns[0]}
		id, e := util.ReadId(s)
		if e != nil {
			id = first
			if !strings.Contains(s, ":") {
				m[db.NAME] = s
			} else {
				r = s
			}
		}
		if i == 0 {
			first = id
		}
		if strings.Contains(r, ":") && db.Contains(db.FILES, bson.M{db.TYPE: project.TEST, db.NAME: strings.Split(r, ":")[1] + ".java", db.SUBID: id}) {
			c, e := userTestChart(id, r)
			if e != nil {
				return nil, e
			}
			d = append(d, c...)
			continue
		}
		m[db.SUBID] = id
		fs, e := db.Files(m, bson.M{db.DATA: 0}, 0, db.TIME)
		if e != nil {
			return nil, e
		}
		c, e := LoadChart(r, fs)
		if e != nil {
			return nil, e
		}
		d = append(d, c...)
	}
	lkeys := make(map[string]bool)
	for _, m := range d {
		for k, v := range m {
			s, ok := v.(string)
			if k == "key" && ok && !lkeys[s] {
				lkeys[s] = true
			}
		}
	}
	return util.JSON(map[string]interface{}{"chart": d})
}

func userTestChart(sid bson.ObjectId, n string) (ChartData, error) {
	ts, e := db.Files(bson.M{db.SUBID: sid, db.NAME: strings.Split(n, ":")[1] + ".java", db.TYPE: project.TEST}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	for _, t := range ts {
		for k, _ := range t.Results {
			if strings.HasPrefix(k, n) {
				return LoadChart(strings.Split(n, ":")[0], []*project.File{t})
			}
		}
	}
	return nil, fmt.Errorf("no results found for %s", n)
}

/*
func srcViewChart(fid bson.ObjectId, result string, subs []string) ([]byte, error) {
	cf, e := db.File(bson.M{db.ID: fid}, bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	c, e := LoadChart(result, []*project.File{cf})
	if e != nil {
		return nil, e
	}
	if len(subs) == 1 {
		return util.JSON(map[string]interface{}{"chart": c})
	}
	fs, e := db.Files(bson.M{db.SUBID: cf.SubId, db.NAME: cf.Name}, bson.M{db.ID: 1}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	var p float64 = -1.0
	for i, f := range fs {
		if f.Id == cf.Id {
			p = float64(i) / float64(len(fs))
			break
		}
	}
	if p < 0 {
		return nil, fmt.Errorf("file %s not found", cf.Id.Hex())
	}
	for _, s := range subs {
		sid, e := util.ReadId(s)
		if e != nil {
			return nil, e
		}
		if sid == cf.SubId {
			continue
		}
		fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: cf.Name}, bson.M{db.DATA: 0}, 0, db.TIME)
		if e != nil || len(fs) == 0 {
			continue
		}
		f, e := determineTest(fs, result, p)
		if e != nil {
			util.Log(e)
			continue
		}
		nc, e := LoadChart(result, []*project.File{f})
		if e != nil {
			return nil, e
		}
		c = append(c, nc...)
	}
	return util.JSON(map[string]interface{}{"chart": c})
}

func testViewChart(fid bson.ObjectId, result, test string, subs []string) ([]byte, error) {
	src, e := db.File(bson.M{db.ID: fid}, bson.M{db.DATA: 0})
	if e != nil {
		return nil, e
	}
	fs, e := db.Files(bson.M{db.SUBID: src.SubId, db.NAME: test}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	c, e := LoadChart(result+"-"+src.Id.Hex(), fs)
	if e != nil {
		return nil, e
	}
	if len(subs) == 1 {
		return util.JSON(map[string]interface{}{"chart": c})
	}
	srcs, e := db.Files(bson.M{db.SUBID: src.SubId, db.NAME: src.Name}, bson.M{db.ID: 1}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	var p float64 = -1.0
	for i, f := range srcs {
		if f.Id == src.Id {
			p = float64(i) / float64(len(srcs))
			break
		}
	}
	if p < 0 {
		return nil, fmt.Errorf("file %s not found", src.Id.Hex())
	}
	for _, s := range subs {
		sid, e := util.ReadId(s)
		if e != nil {
			return nil, e
		}
		if sid == src.SubId {
			continue
		}
		fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: test}, bson.M{db.DATA: 0}, 0, db.TIME)
		if e != nil || len(fs) == 0 {
			continue
		}
		srcs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: src.Name}, bson.M{db.ID: 1}, 0, db.TIME)
		if e != nil || len(srcs) == 0 {
			continue
		}
		src, e := determineSrc(srcs, result, test, p)
		nc, e := LoadChart(result+"-"+src.Id.Hex(), fs)
		if e != nil {
			return nil, e
		}
		c = append(c, nc...)
	}
	return util.JSON(map[string]interface{}{"chart": c})
}

func determineSrc(fs []*project.File, r, n string, p float64) (*project.File, error) {
	i := int(float64(len(fs)) * p)
	j := 0
	for i+j >= 0 && i+j < len(fs) {
		rc, e := db.Count(db.FILES, bson.M{db.NAME: n, db.SUBID: fs[i+j].SubId, db.RESULTS + "." + r + "-" + fs[i+j].Id.Hex(): bson.M{db.EXISTS: true}})
		if e != nil {
			return nil, e
		}
		if rc > 0 {
			return fs[i+j], nil
		}
		if (j < 0 && i-j+1 < len(fs)) || (j > 0 && i-j >= 0) {
			j = -j
		} else if j < 0 {
			j--
		}
		if j >= 0 {
			j++
		}
	}
	return nil, errors.New("no result file found")
}

func determineTest(fs []*project.File, r string, p float64) (*project.File, error) {
	i := int(float64(len(fs)) * p)
	j := 0
	for i+j >= 0 && i+j < len(fs) {
		rc, e := db.Count(db.RESULTS, bson.M{db.TYPE: r, db.FILEID: fs[i+j].Id})
		if e != nil {
			return nil, e
		}
		if rc > 0 {
			return fs[i+j], nil
		}
		if (j < 0 && i-j+1 < len(fs)) || (j > 0 && i-j >= 0) {
			j = -j
		} else if j < 0 {
			j--
		}
		if j >= 0 {
			j++
		}
	}
	return nil, errors.New("no result file found")
}
*/
