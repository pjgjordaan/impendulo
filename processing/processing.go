package processing

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
)

const (
	INCOMPLETE = "incomplete.gob"
	TESTS      = "tests"
	SRC        = "src"
)


//testBuilder is used to setup a project's tests.
type testBuilder struct {
	tests   map[string]bool
	m       *sync.Mutex
	testDir string
}

//newTestBuilder
func newTestBuilder() *testBuilder {
	dir := filepath.Join(os.TempDir(), TESTS)
	return &testBuilder{make(map[string]bool), new(sync.Mutex), dir}
}


//setup extracts a project's tests from db to filesystem for execution.
//It returns true if this was successful.
func (this *testBuilder) setup(project string) bool {
	this.m.Lock()
	defer this.m.Unlock()
	if this.tests[project] {
		return true
	}
	tests, ok := getTests(project)
	if !ok {
		return false
	}
	err := util.Unzip(filepath.Join(this.testDir, project), tests.Data)
	if err != nil {
		util.Log(err)
		return false
	}
	this.tests[project] = true
	return true
}

//getTests retrieves project tests from database.
//It returns true if this was successful.
func getTests(project string) (*submission.File, bool) {
	smatcher := bson.M{submission.PROJECT: project, submission.MODE: submission.TEST_MODE}
	s, err := db.GetSubmission(smatcher)
	if err != nil {
		util.Log("Error retrieving test submission:", err)
		return nil, false
	}
	fmatcher := bson.M{submission.SUBID: s.Id}
	f, err := db.GetFile(fmatcher)
	if err != nil {
		util.Log(err)
		return nil, false
	}
	return f, true
}

//Serve spawns new processing routines for each new submission received on subChan.
//New files are received on fileChan and then sent to the relevant submission process.
//Incomplete submissions are read from disk and reprocessed using the processStored function.   
func Serve(subChan chan *submission.Submission, fileChan chan *submission.File) {
	// Start handlers
	busy := make(chan bson.ObjectId)
	done := make(chan bson.ObjectId)
	go statusListener(busy, done)
	builder := newTestBuilder()
	go func() {
		stored := getStored()
		for subId, _ := range stored {
			processStored(subId, busy, done, builder)
		}
	}()
	subs := make(map[bson.ObjectId]chan *submission.File)
	for {
		select {
		case sub := <-subChan:
			subs[sub.Id] = make(chan *submission.File)
			go processNew(sub, subs[sub.Id], busy, done, builder)
		case file := <-fileChan:
			if ch, ok := subs[file.SubId]; ok {
				ch <- file
			} else {
				util.Log(fmt.Errorf("No channel found for submission: %q", file.SubId))
			}
		}
	}
}

//getStored retrieves incompletely processed submissions from the filesystem.
func getStored() map[bson.ObjectId]bool {
	stored, err := util.LoadMap(INCOMPLETE)
	if err != nil {
		util.Log(err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}

//saveActive saves active submissions to the filesystem.
func saveActive(active map[bson.ObjectId]bool) {
	err := util.SaveMap(active, INCOMPLETE)
	if err != nil {
		util.Log(err)
	}
}

//statusListener listens for new submissions and adds them to the map of active processes. 
//It also listens for completed submissions and removes them from the active process map.
//Finally it detects Kill and Interrupt signals, saving the active processes if they are detected.
func statusListener(busy, done chan bson.ObjectId) {
	active := getStored()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Kill, os.Interrupt)
	for {
		select {
		case id := <-busy:
			active[id] = true
		case id := <-done:
			delete(active, id)
		case <-quit:
			saveActive(active)
			os.RemoveAll(filepath.Join(os.TempDir(), TESTS))
			os.Exit(0)
		}
	}
}

//processStored processes an incompletely processed submission. 
//It retrieves all files in the submission from the db and processes them. 
func processStored(subId bson.ObjectId, busy, done chan bson.ObjectId, builder *testBuilder) {
	busy <- subId
	defer os.RemoveAll(filepath.Join(os.TempDir(), subId.Hex()))
	sub, err := db.GetSubmission(bson.M{submission.SUBID: subId})
	if err != nil {
		util.Log(err)
		return
	}
	count := 0
	hasTests := builder.setup(sub.Project)
	for {
		matcher := bson.M{submission.SUBID: subId, submission.NUM: count}
		file, err := db.GetFile(matcher)
		if err != nil {
			util.Log(err)
			return
		}
		err = processFile(file, hasTests)
		if err != nil {
			util.Log(err)
			return
		}
	}
	done <- subId
}

//processNew processes a new submission.
//It listens for incoming files on fileChan and processes them.
func processNew(sub *submission.Submission, fileChan chan *submission.File, busy, done chan bson.ObjectId, builder *testBuilder) {
	busy <- sub.Id
	defer os.RemoveAll(filepath.Join(os.TempDir(), sub.Id.Hex()))
	hasTests := builder.setup(sub.Project)
	for file := range fileChan {
		err := processFile(file, hasTests)
		if err != nil {
			util.Log(err)
			return
		}
	}
	done <- sub.Id
}

//processFile processes a file according to its type. 
//If an error occurs, it is returned.
func processFile(f *submission.File, hasTests bool) error {
	util.Log("Processing file: ", f.Id)
	t := f.Type()
	if t == submission.ARCHIVE {
		err := processArchive(f, hasTests)
		if err != nil {
			return err
		}
		db.RemoveFileByID(f.Id)
	} else if t == submission.SRC {
		err := evaluate(f, compile, hasTests)
		if err != nil {
			return err
		}
	} else if t == submission.EXEC {
		err := evaluate(f, alreadyCompiled, hasTests)
		if err != nil {
			return err
		}
	}
	util.Log("Processed file: ", f.Id)
	return nil
}

//processArchive extracts files from an archive and processes them.
func processArchive(archive *submission.File, hasTests bool) error {
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		return err
	}
	for name, data := range files {
		info, err := submission.ParseName(name)
		if err != nil {
			return err
		}
		matcher := bson.M{submission.INFO: info}
		f, err := db.GetFile(matcher)
		if err != nil {
			f = submission.NewFile(archive.SubId, info, data)
			err = db.AddFile(f)
			if err != nil {
				return err
			}
		}
		err = processFile(f, hasTests)
		if err != nil {
			return err
		}
	}
	return nil
}

//evaluate evaluates a source or compiled file by attempting to run tests and tools on it.
func evaluate(orig *submission.File, comp compFunc, hasTests bool) error {
	ti, err := extractFile(orig)
	if err != nil {
		return err
	}
	compiled, err := comp(orig.Id, ti)
	if err != nil {
		return err
	}
	if !compiled {
		return nil
	}
	f, err := db.GetFile(bson.M{submission.ID: orig.Id})
	if err != nil {
		return err
	}
	util.Log("Evaluating: ", f.Id, f.Info)
	if hasTests {
		err = runTests(f, ti)
		if err != nil {
			return err
		}
		util.Log("Tested: ", f.Id)
	}
	err = runTools(f, ti)
	if err != nil {
		return err
	}
	util.Log("Ran tools: ", f.Id)
	return nil
}

//extractFile saves a file to filesystem.
//It returns file info used by tools & tests.
func extractFile(f *submission.File) (*tool.TargetInfo, error) {
	matcher := bson.M{submission.ID: f.SubId}
	s, err := db.GetSubmission(matcher)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(os.TempDir(), f.SubId.Hex())
	ti := tool.NewTarget(s.Project, f.InfoStr(submission.NAME), s.Lang, f.InfoStr(submission.PKG), dir)
	err = util.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), f.Data)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

type compFunc func(fileId bson.ObjectId, ti *tool.TargetInfo) (bool, error)

//compile compiles a java source file and writes the results thereof to the database.
//It returns true if compiled successfully.
func compile(fileId bson.ObjectId, ti *tool.TargetInfo) (bool, error) {
	comp, err := db.GetTool(ti.GetCompiler())
	if err != nil {
		return false, err
	}
	res, err := tool.RunTool(fileId, ti, comp, map[string]string{tool.CP: ti.Dir})
	if err != nil {
		return false, err
	}
	err = addResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}

//alreadyCompiled stores compilation info for compiled files in the db.
func alreadyCompiled(fileId bson.ObjectId, ti *tool.TargetInfo) (bool, error) {
	comp, err := db.GetTool(ti.GetCompiler())
	if err != nil {
		return false, err
	}
	res := tool.NewResult(fileId, comp.Id, comp.Name, comp.OutName, comp.ErrName, []byte(""), []byte(""), nil)
	err = addResult(res)
	if err != nil {
		return false, err
	}
	return true, nil
}

//runTests sets up and runs tests on executable file. 
//If an error is encountered, iteration stops and the error is returned.
func runTests(f *submission.File, ti *tool.TargetInfo) error {
	tests := []string{"EasyTests", "AllTests"}
	for _, test := range tests {
		dir := filepath.Join(os.TempDir(), TESTS, ti.Project, SRC)
		compiled, err := compileTest(f, ti, test, dir)
		if err != nil {
			return err
		}
		if !compiled {
			continue
		}
		err = runTest(f, ti, test, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

//compileTest
func compileTest(f *submission.File, ti *tool.TargetInfo, test, dir string) (bool, error) {
	if _, ok := f.Results[test+"_"+tool.COMPILE]; ok {
		return true, nil
	}
	//Run if not run previously
	res, err := tool.CompileTest(f.Id, ti, test, dir)
	if err != nil {
		return false, err
	}
	err = addResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}

//runTest
func runTest(f *submission.File, ti *tool.TargetInfo, test, dir string) error {
	if _, ok := f.Results[test+"_"+tool.RUN]; ok {
		return nil
	}
	//Run if not run previously
	res, err := tool.RunTest(f.Id, ti, test, dir)
	if err != nil {
		return err
	}
	return addResult(res)
}


//runTools runs all available tools on a file, skipping previously run tools.
func runTools(f *submission.File, ti *tool.TargetInfo) error {
	all, err := db.GetTools(bson.M{tool.LANG: ti.Lang})
	if err != nil {
		return err
	}
	for _, t := range all {
		if _, ok := f.Results[t.Name]; ok {
			continue
		}
		res, err := tool.RunTool(f.Id, ti, t, map[string]string{})
		if err != nil {
			return err
		}
		err = addResult(res)
		if err != nil {
			return err
		}
	}
	return nil
}


//addResult adds a tool result to the db.
//It updates the associated file's list of results to point to this new result.
func addResult(res *tool.Result) error {
	matcher := bson.M{submission.ID: res.FileId}
	change := bson.M{db.SET: bson.M{submission.RES + "." + res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}

//init sets up tool configuration for testing
func init() {
	_, err := db.GetTool(bson.M{tool.NAME: "findbugs"})
	if err != nil {
		fb := &tool.Tool{bson.NewObjectId(), "findbugs", tool.JAVA, "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", tool.WARNS, []string{tool.JAVA, "-jar"}, []string{"-textui", "-low"}, bson.M{}, tool.PKG_PATH}
		db.AddTool(fb)
	}
	_, err = db.GetTool(bson.M{tool.NAME: tool.COMPILE})
	if err != nil {
		javac := &tool.Tool{bson.NewObjectId(), tool.COMPILE, tool.JAVA, tool.JAVAC, tool.WARNS, tool.ERRS, []string{}, []string{"-implicit:class"}, bson.M{tool.CP: ""}, tool.FILE_PATH}
		db.AddTool(javac)
	}
}
