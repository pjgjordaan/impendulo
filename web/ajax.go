package web

import (
	"code.google.com/p/gorilla/pat"

	"errors"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
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
		"chart": getChart, "tools": getTools, "users": getUsers,
		"skeletons": getSkeletons, "code": getCode, "submissions": submissions,
	}
	for n, f := range fs {
		r.Add("GET", "/"+n, f)
	}
}

func submissions(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	s, e := db.Submissions(bson.M{db.PROJECTID: pid}, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"submissions": s})
}

func getUsers(r *http.Request) ([]byte, error) {
	pid, _, e := getProjectId(r)
	if e != nil {
		return nil, e
	}
	u, e := users(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

func getTools(r *http.Request) ([]byte, error) {
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

func getCode(r *http.Request) ([]byte, error) {
	rid, _, e := getId(r, "resultid", "result")
	if e != nil {
		return nil, e
	}
	tr, e := db.ToolResult(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
	if e != nil {
		return nil, e
	}
	f, e := db.File(bson.M{db.ID: tr.GetFileId()}, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
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
	if f, e := GetStrings(r, "testfileid"); e == nil && len(f) == 1 {
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
	}
	var d ChartData
	for _, s := range subs {
		id, e := util.ReadId(s)
		if e != nil {
			return nil, e
		}
		fs, e := db.Files(bson.M{db.SUBID: id, db.NAME: ns[0]}, bson.M{db.DATA: 0}, db.TIME)
		if e != nil {
			return nil, e
		}
		c, e := LoadChart(rs[0], fs)
		if e != nil {
			return nil, e
		}
		d = append(d, c...)
	}
	return util.JSON(map[string]interface{}{"chart": d})
}

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
	fs, e := db.Files(bson.M{db.SUBID: cf.SubId, db.NAME: cf.Name}, bson.M{db.ID: 1}, db.TIME)
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
		fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: cf.Name}, bson.M{db.DATA: 0}, db.TIME)
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
	fs, e := db.Files(bson.M{db.SUBID: src.SubId, db.NAME: test}, bson.M{db.DATA: 0}, db.TIME)
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
	srcs, e := db.Files(bson.M{db.SUBID: src.SubId, db.NAME: src.Name}, bson.M{db.ID: 1}, db.TIME)
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
		fs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: test}, bson.M{db.DATA: 0}, db.TIME)
		if e != nil || len(fs) == 0 {
			continue
		}
		srcs, e := db.Files(bson.M{db.SUBID: sid, db.NAME: src.Name}, bson.M{db.ID: 1}, db.TIME)
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
