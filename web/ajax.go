package web

import (
	"code.google.com/p/gorilla/pat"

	"time"

	"errors"

	"sort"
	"strings"

	"fmt"

	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/processor/status"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/jacoco"
	"github.com/godfried/impendulo/tool/jpf"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/tool/pmd"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"
	"github.com/godfried/impendulo/web/charts"
	"github.com/godfried/impendulo/web/context"
	"github.com/godfried/impendulo/web/webutil"
	"labix.org/v2/mgo/bson"

	"net/http"
)

type (
	AJAXGet  func(*http.Request) ([]byte, error)
	AJAXPost func(http.ResponseWriter, *http.Request) error
	Select   struct {
		Id, Name string
		User     bool
	}
	Selects []*Select
)

var (
	CountsError   = errors.New("unsupported counts request")
	CommentsError = errors.New("unsupported comments request")
	ResultsError  = errors.New("cannot retrieve results")
)

func (s Selects) Len() int {
	return len(s)
}

func (s Selects) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Selects) Less(i, j int) bool {
	return strings.ToLower(s[i].Name) <= strings.ToLower(s[j].Name)
}

func (a AJAXGet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, e := a(r)
	if e != nil {
		util.Log(e, LOG_HANDLERS)
		b, _ = util.JSON(map[string]interface{}{"error": e.Error()})
	}
	fmt.Fprint(w, string(b))
}

func (a AJAXPost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := a(w, r); e != nil {
		b, _ := util.JSON(map[string]interface{}{"error": e.Error()})
		util.Log(e, LOG_HANDLERS)
		fmt.Fprint(w, string(b))
	}
}

func GenerateAJAX(r *pat.Router) {
	gets := map[string]AJAXGet{
		"chart": getChart, "usernames": getUsernames, "collections": collections, "pmdrules": ajaxRules,
		"skeletons": getSkeletons, "submissions": submissions, "results": ajaxResults, "jpflisteners": ajaxListeners,
		"langs": getLangs, "projects": getProjects, "files": ajaxFiles, "tools": ajaxTools, "jpfsearches": ajaxSearches,
		"code": ajaxCode, "users": ajaxUsers, "permissions": ajaxPerms, "comparables": ajaxComparables,
		"tests": ajaxTests, "test-types": testTypes, "filenames": fileNames, "status": ajaxStatus, "counts": ajaxCounts,
		"comments": ajaxComments,
	}
	for n, f := range gets {
		r.Add("GET", "/"+n, f)
	}
	posts := map[string]AJAXPost{
		"setcontext": ajaxSetContext, "addcomment": addComment,
	}
	for n, f := range posts {
		r.Add("POST", "/"+n, f)
	}
}

func ajaxStatus(r *http.Request) ([]byte, error) {
	type wrapper struct {
		s *status.S
		e error
	}
	sc := make(chan wrapper)
	go func() {
		s, e := mq.GetStatus()
		w := wrapper{s, e}
		sc <- w
	}()
	select {
	case <-time.After(15 * time.Second):
		return util.JSON(map[string]interface{}{"status": status.New()})
	case w := <-sc:
		if w.e != nil {
			util.Log(w.e)
			return util.JSON(map[string]interface{}{"status": status.New()})
		}
		return util.JSON(map[string]interface{}{"status": w.s})
	}
}

func ajaxComments(r *http.Request) ([]byte, error) {
	c, e := commentor(r)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"comments": c.LoadComments()})
}

func commentor(r *http.Request) (project.Commentor, error) {
	if id, e := convert.Id(r.FormValue("file-id")); e == nil {
		return db.File(bson.M{db.ID: id}, bson.M{db.COMMENTS: 1})
	} else if id, e := convert.Id(r.FormValue("submission-id")); e == nil {
		return db.Submission(bson.M{db.ID: id}, bson.M{db.COMMENTS: 1})
	}
	return nil, CommentsError
}

func fileNames(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return nil, e
	}
	ns, e := db.ProjectFileNames(pid)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"filenames": ns})
}

func ajaxSetContext(w http.ResponseWriter, r *http.Request) error {
	c, e := loadContext(r)
	if e != nil {
		return e
	}
	if e := c.Browse.Update(r); e != nil {
		return e
	}
	return c.Save(r, w)
}

func addComment(w http.ResponseWriter, r *http.Request) error {
	c, e := loadContext(r)
	if e != nil {
		return e
	}
	u, e := c.Username()
	if e != nil {
		return e
	}
	fid, e := convert.Id(r.FormValue("File-id"))
	if e != nil {
		return e
	}
	start, e := convert.Int(r.FormValue("Start"))
	if e != nil {
		return e
	}
	end, e := convert.Int(r.FormValue("End"))
	if e != nil {
		return e
	}
	d, e := webutil.String(r, "Data")
	if e != nil {
		return e
	}
	f, e := db.File(bson.M{db.ID: fid}, bson.M{db.COMMENTS: 1})
	if e != nil {
		return e
	}
	f.Comments = append(f.Comments, &project.Comment{Data: d, User: u, Start: start, End: end})
	return db.Update(db.FILES, bson.M{db.ID: fid}, bson.M{db.SET: bson.M{db.COMMENTS: f.Comments}})
}

func testTypes(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"types": junit.TestTypes()})
}

func ajaxCounts(r *http.Request) ([]byte, error) {
	if sid, e := convert.Id(r.FormValue("submission-id")); e == nil {
		return submissionCounts(sid)
	}
	return nil, CountsError
}

func submissionCounts(sid bson.ObjectId) ([]byte, error) {
	all, e := db.Count(db.FILES, bson.M{db.SUBID: sid})
	if e != nil {
		return nil, e
	}
	s, e := db.Count(db.FILES, bson.M{db.SUBID: sid, db.TYPE: project.SRC})
	if e != nil {
		return nil, e
	}
	l, e := db.Count(db.FILES, bson.M{db.SUBID: sid, db.TYPE: project.LAUNCH})
	if e != nil {
		return nil, e
	}
	t, e := db.Count(db.FILES, bson.M{db.SUBID: sid, db.TYPE: project.TEST})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"counts": map[string]int{"all": all, "source": s, "launch": l, "test": t}})
}

func ajaxListeners(r *http.Request) ([]byte, error) {
	l, e := jpf.Listeners()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"listeners": l})
}

func ajaxSearches(r *http.Request) ([]byte, error) {
	s, e := jpf.Searches()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"searches": s})
}

func ajaxRules(r *http.Request) ([]byte, error) {
	rs, e := pmd.RuleSet()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"rules": rs})
}

//ajaxResults retrieves the names  of all results found within a particular
//project or by a particular user.
func ajaxResults(r *http.Request) ([]byte, error) {
	var rs []string
	if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
		rs = db.ProjectResults(pid)
	} else if u, e := webutil.String(r, "user-id"); e == nil {
		rs = db.UserResults(u)
	} else {
		return nil, ResultsError
	}
	s := make(Selects, len(rs))
	for i, r := range rs {
		s[i] = &Select{Id: r, Name: strings.Replace(r, ":", " \u2192 ", -1)}
	}
	sort.Sort(s)
	return util.JSON(map[string]interface{}{"results": s})
}

//ajaxComparables retrieves other results which a given result
//can be compared to, i.e. different unit tests.
func ajaxComparables(r *http.Request) ([]byte, error) {
	id, e := convert.Id(r.FormValue("id"))
	if e != nil {
		return nil, e
	}
	tr, e := db.ToolResult(bson.M{db.ID: id}, nil)
	if e != nil {
		return nil, e
	}
	if tr.GetType() != jacoco.NAME && tr.GetType() != junit.NAME {
		return util.JSON(map[string]interface{}{"comparables": []*Select{}})
	}
	f, e := db.File(bson.M{db.ID: tr.GetFileId()}, bson.M{db.SUBID: 1})
	if e != nil {
		return nil, e
	}
	s, e := db.Submission(bson.M{db.ID: f.SubId}, bson.M{db.PROJECTID: 1})
	if e != nil {
		return nil, e
	}
	ts, e := db.JUnitTests(bson.M{db.PROJECTID: s.ProjectId, db.NAME: bson.M{db.NE: tr.GetName() + ".java"}, db.TYPE: bson.M{db.NE: junit.USER}}, bson.M{db.NAME: 1, db.TYPE: 1})
	if e != nil {
		return nil, e
	}
	m := bson.M{db.SUBID: f.SubId, db.TYPE: project.TEST}
	if ut, e := db.File(bson.M{db.ID: tr.GetTestId()}, bson.M{db.NAME: 1}); e == nil {
		m[db.ID] = bson.M{db.NE: ut.Id}
	} else if !db.Contains(db.TESTS, bson.M{db.ID: tr.GetTestId()}) {
		return nil, e
	}
	uts, e := db.Files(m, bson.M{db.DATA: 0}, 0, "-"+db.TIME)
	if e != nil {
		return nil, e
	}
	cmp := make([]*Select, len(ts)+len(uts))
	for i, t := range ts {
		n, _ := util.Extension(t.Name)
		cmp[i] = &Select{tr.GetType() + ":" + n, tr.GetType() + " \u2192 " + n, false}
	}
	for i, ut := range uts {
		n, _ := util.Extension(ut.Name)
		rd, e := context.NewResult(tr.GetType() + ":" + n + "-" + ut.Id.Hex())
		if e != nil {
			return nil, e
		}
		cmp[i+len(ts)] = &Select{rd.Raw(), rd.Format(), true}
	}
	return util.JSON(map[string]interface{}{"comparables": cmp})
}

//ajaxPerms retrieves the different user permission levels.
func ajaxPerms(r *http.Request) ([]byte, error) {
	return util.JSON(map[string]interface{}{"permissions": user.PermissionInfos()})
}

//ajaxUsers retrieves a list of users.
func ajaxUsers(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if n, e := webutil.String(r, "name"); e == nil {
		m[db.ID] = n
	}
	u, e := db.Users(m)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"users": u})
}

//ajaxCode loads code for a given src file or test.
func ajaxCode(r *http.Request) ([]byte, error) {
	if tid, e := convert.Id(r.FormValue("test-id")); e == nil {
		t, e := db.JUnitTest(bson.M{db.ID: tid}, bson.M{db.TEST: 1})
		if e != nil {
			return nil, e
		}
		return util.JSON(map[string]interface{}{"code": string(t.Test)})
	}
	m := bson.M{}
	if rid, e := convert.Id(r.FormValue("result-id")); e == nil {
		tr, e := db.ToolResult(bson.M{db.ID: rid}, bson.M{db.FILEID: 1})
		if e != nil {
			return nil, e
		}
		m[db.ID] = tr.GetFileId()
	}
	if id, e := convert.Id(r.FormValue("file-id")); e == nil {
		m[db.ID] = id
	}
	if n, e := webutil.String(r, "tool-name"); e == nil {
		d, e := context.NewResult(n)
		if e != nil {
			return nil, e
		}
		if d.FileID != "" {
			m[db.ID] = d.FileID
		} else if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
			return loadTestCode(pid, d.Name)
		} else {
			return nil, fmt.Errorf("could not load code for %s", d.Format())
		}
	}
	f, e := db.File(m, bson.M{db.DATA: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(f.Data)})
}

//loadTestCode loads a unit test's code
func loadTestCode(pid bson.ObjectId, n string) ([]byte, error) {
	t, e := db.JUnitTest(bson.M{db.PROJECTID: pid, db.NAME: n + ".java"}, bson.M{db.TEST: 1})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"code": string(t.Test)})
}

//collections retrieves the names of all collections in the current database.
func collections(r *http.Request) ([]byte, error) {
	n, e := webutil.String(r, "db")
	if e != nil {
		return nil, e
	}
	c, e := db.Collections(n)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"collections": c})
}

//getProjects loads a list of projects.
func getProjects(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = pid
	}
	p, e := db.Projects(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"projects": p})
}

//ajaxTools loads a list of available tools for a given project.
func ajaxTools(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
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
	if sid, e := convert.Id(r.FormValue("submission-id")); e == nil {
		m[db.SUBID] = sid
	}
	if id, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	format, _ := webutil.String(r, "format")
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
	if sid, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = sid
	}
	if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
		m[db.PROJECTID] = pid
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"submissions": s})
}

func getUsernames(r *http.Request) ([]byte, error) {
	pid, e := convert.Id(r.FormValue("project-id"))
	var u []string
	if e != nil {
		u, e = db.Usernames(nil)
	} else {
		u, e = db.ProjectUsernames(pid)
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
	pid, e := convert.Id(r.FormValue("project-id"))
	if e != nil {
		return nil, e
	}
	s, e := db.Skeletons(bson.M{db.PROJECTID: pid}, bson.M{db.DATA: 0}, db.NAME)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"skeletons": s})
}

func ajaxTests(r *http.Request) ([]byte, error) {
	m := bson.M{}
	if pid, e := convert.Id(r.FormValue("project-id")); e == nil {
		m[db.PROJECTID] = pid
	}
	if id, e := convert.Id(r.FormValue("id")); e == nil {
		m[db.ID] = id
	}
	t, e := db.JUnitTests(m, bson.M{db.TEST: 0, db.DATA: 0})
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"tests": t})
}

func getChart(r *http.Request) ([]byte, error) {
	if e := r.ParseForm(); e != nil {
		return nil, e
	}
	t, e := webutil.String(r, "type")
	if e != nil {
		return nil, e
	}
	switch t {
	case "file":
		return fileChart(r)
	case "submission":
		return submissionChart(r)
	case "overview":
		return overviewChart(r)
	default:
		return nil, fmt.Errorf("unsupported chart type %s", t)
	}
}

func overviewChart(r *http.Request) ([]byte, error) {
	v, e := webutil.String(r, "view")
	if e != nil {
		return nil, e
	}
	var f func() (charts.D, error)
	switch v {
	case "user":
		f = charts.User
	case "project":
		f = charts.Project
	default:
		return nil, fmt.Errorf("unknown view %s", v)
	}
	c, e := f()
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart": c})
}

func submissionChart(r *http.Request) ([]byte, error) {
	rn, e := webutil.String(r, "result")
	if e != nil {
		return nil, e
	}
	rd, e := context.NewResult(rn)
	if e != nil {
		return nil, e
	}
	m := bson.M{}
	t, e := webutil.String(r, "submission-type")
	if e != nil {
		return nil, e
	}
	id, e := webutil.String(r, "id")
	if e != nil {
		return nil, e
	}
	sc, e := webutil.String(r, "score")
	if e != nil {
		return nil, e
	}
	switch t {
	case "project":
		pid, e := convert.Id(id)
		if e != nil {
			return nil, e
		}
		m[db.PROJECTID] = pid
	case "user":
		m[db.USER] = id
	default:
		return nil, fmt.Errorf("invalid submission chart type %s", t)
	}
	s, e := db.Submissions(m, nil)
	if e != nil {
		return nil, e
	}
	c, e := charts.Submission(s, rd, sc)
	if e != nil {
		return nil, e
	}
	return util.JSON(map[string]interface{}{"chart": c})
}

func fileChart(r *http.Request) ([]byte, error) {
	sid, e := convert.Id(r.FormValue("submission-id"))
	if e != nil {
		return nil, e
	}
	fn, e := webutil.String(r, "file")
	if e != nil {
		return nil, e
	}
	rn, e := webutil.String(r, "result")
	if e != nil {
		return nil, e
	}
	rd, e := context.NewResult(rn)
	if e != nil {
		return nil, e
	}
	subs, e := webutil.Strings(r, "submissions[]")
	if e != nil {
		return nil, e
	}
	cmps, e := webutil.Strings(r, "comparables[]")
	if e != nil {
		return nil, e
	}
	var d charts.D
	for _, s := range subs {
		if c, e := _fileChart(s, fn, rd); e != nil {
			util.Log(e)
		} else {
			d = append(d, c...)
		}
	}
	for _, cmp := range cmps {
		if c, e := _cmpChart(sid, cmp, fn); e != nil {
			util.Log(e)
		} else {
			d = append(d, c...)
		}
	}
	return util.JSON(map[string]interface{}{"chart": d})
}

func _fileChart(s, fn string, r *context.Result) (charts.D, error) {
	id, e := convert.Id(s)
	if e != nil {
		return nil, e
	} else if r.FileID != "" {
		r.FileID = findTestId(id)
	}
	fs, e := db.Files(bson.M{db.NAME: fn, db.SUBID: id}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	return charts.Tool(r, fs)
}

func _cmpChart(sid bson.ObjectId, cmp, fn string) (charts.D, error) {
	r, e := context.NewResult(cmp)
	if e != nil {
		return nil, e
	}
	fs, e := db.Files(bson.M{db.NAME: fn, db.SUBID: sid}, bson.M{db.DATA: 0}, 0, db.TIME)
	if e != nil {
		return nil, e
	}
	return charts.Tool(r, fs)
}

func findTestId(sid bson.ObjectId) bson.ObjectId {
	ts, e := db.Files(bson.M{db.SUBID: sid, db.TYPE: project.TEST}, bson.M{db.ID: 1}, 0, "-"+db.TIME)
	if e != nil {
		return ""
	}
	for _, t := range ts {
		if db.Contains(db.RESULTS, bson.M{db.TESTID: t.Id}) {
			return t.Id
		}
	}
	return ""
}
