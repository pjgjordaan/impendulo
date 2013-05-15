package processing

import (
	"labix.org/v2/mgo/bson"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/util"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
)

const (
	INCOMPLETE = "incomplete.gob"
	TESTS = "tests"
	SRC = "src"
)

var builder *testBuilder

type testBuilder struct {
	tests   map[string]bool
	m       *sync.Mutex
	testDir string
}

func newTestBuilder() *testBuilder {
	dir := filepath.Join(os.TempDir(), TESTS)
	return &testBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

/*
Extracts project tests from db to filesystem for execution.
*/
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
		util.Log("Error extracting tests:", project, err)
		return false
	} 
	this.tests[project] = true
	return true
}

/*
Retrieves project tests from database.
*/
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
		util.Log("Error retrieving test files:", err)
		return nil, false
	}
	return f, true
}

/*
Spawns new processing routines for each file which needs to be processed.
*/
func Serve(fileChan chan *submission.File) {
	builder = newTestBuilder()
	// Start handlers
	busy := make(chan bson.ObjectId)
	done := make(chan bson.ObjectId)
	go statusListener(busy, done)
	go func() {
		stored := getStored()
		for subId, _ := range stored {
			processStored(subId, busy, done)
		}
	}()
	subs := make(map[bson.ObjectId]chan *submission.File)
	for file := range fileChan {
		if ch, ok := subs[file.SubId]; ok {
			ch <- file
		} else {
			subs[file.SubId] = make(chan *submission.File)
			go processNew(file.SubId, subs[file.SubId], busy, done)
			subs[file.SubId] <- file
		}
	}
}

/*
Retrieves incompletely processed submissions from  the filesystem.
*/
func getStored()  map[bson.ObjectId]bool {
	stored, err := util.LoadMap(INCOMPLETE)
	if err != nil{
		util.Log("Unable to read stored map: ", err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}


/*
Retrieves incompletely processed submissions from  the filesystem.
*/
func saveActive(active map[bson.ObjectId]bool) {
	err := util.SaveMap(active, INCOMPLETE)
	if err != nil{
		util.Log("Unable to save active processes map: ", err)
	}
}

/*
Listens for new submissions and adds them to the map of active processes. Listens for completed submissions and 
removes them from the active process map. Listens for Kill or Interrupt signals and saves the active processes if
these signals are detected. 
*/
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
		case sig := <-quit:
			util.Log("Received interrupt signal: ", sig)
			saveActive(active)
			os.RemoveAll(filepath.Join(os.TempDir(), TESTS))
			os.Exit(0)
		}
	}
}

/*
Processes an incomplete submission. Retrieves all files in the submission from the db
and processes them. 
*/
func processStored(subId bson.ObjectId, busy, done chan bson.ObjectId) {
	busy <- subId
	defer os.RemoveAll(filepath.Join(os.TempDir(), subId.Hex()))
	count := 0
	for {
		matcher := bson.M{submission.SUBID: subId, submission.NUM: count}
		f, err := db.GetFile(matcher)
		if err != nil {
			util.Log("Error retrieving file:", err)
			return
		}
		processFile(f)
	}
	done <- subId
}


/*
Processes a new submission. Listens for incoming files and processes them.
*/
func processNew(subId bson.ObjectId, fileChan chan *submission.File, busy, done chan bson.ObjectId) {
	busy <- subId
	for file := range fileChan {
		processFile(file)
	}
	err := os.RemoveAll(filepath.Join(os.TempDir(), subId.Hex()))
	if err != nil {
		util.Log("Error cleaning files:", err)
	}
	done <- subId
}


/*
Processes a file according to its type.
*/
func processFile(f *submission.File) {
	util.Log("Processing file: ", f.Id)
	t := f.Type()
	if t == submission.ARCHIVE {
		processArchive(f)
		db.RemoveFileByID(f.Id)
	} else if t == submission.SRC {
		evaluate(f, compile)
	} else if t == submission.EXEC {
		evaluate(f, alreadyCompiled)
	}
	util.Log("Processed file: ", f.Id)
}

/*
 Evaluates a submitted file (source or compiled) by
 attempting to run tests and tools on it.
*/
func evaluate(orig *submission.File, comp compFunc) {
	f, ti, ok := setupFile(orig, comp)
	if ok {
		util.Log("Evaluating: ", f.Id, f.Info)
		runTests(f, ti)
		util.Log("Tested: ", f.Id)
		runTools(f, ti)
		util.Log("Ran tools: ", f.Id)
	}
}

func setupFile(f *submission.File, comp compFunc)(*submission.File, *tool.TargetInfo, bool){
	ti, ok := extractFile(f)
	if !ok {
		return nil, nil, false
	}
	if !comp(f.Id, ti){
		return nil, nil, false
	}
	matcher := bson.M{submission.ID: f.Id}
	f, err := db.GetFile(matcher)
	if err != nil{
		util.Log("Error retrieving file", err)
		return nil, nil, false
	}
	util.Log("Compiled: ", f.Id)
	return f, ti, true
}

/*
 Saves file to filesystem. Returns file info used by tools & tests.
*/
func extractFile(f *submission.File) (*tool.TargetInfo, bool) {
	matcher := bson.M{submission.ID:f.SubId}
	s, err := db.GetSubmission(matcher)
	if err != nil {
		util.Log("Error retrieving submission",err)
		return nil, false
	}
	dir := filepath.Join(os.TempDir(), f.SubId.Hex())
	ti := tool.NewTarget(s.Project, f.InfoStr(submission.NAME), s.Lang, f.InfoStr(submission.PKG), dir)
	err = util.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), f.Data)
	if err != nil {
		util.Log("Error saving file: ", f.Id, err)
		return nil, false
	}
	return ti, true
}

/*
 Sets up and runs test on executable file.
*/
func runTests(f *submission.File, ti *tool.TargetInfo) {
	tests := []string{"EasyTests", "AllTests"}
	ok := builder.setup(ti.Project)
	if !ok{
		util.Log("No tests found for ", ti.Project)
		return
	}
	for _, test := range tests {
		if _, ok := f.Results[test+"_compile"]; ok {
			continue
		}
		//Run if not run previously
		dir := filepath.Join(os.TempDir(), TESTS, ti.Project, SRC)
		compRes, compErr := tool.CompileTest(f.Id, ti, test, dir)
		dbErr := addResult(compRes)
		if dbErr != nil{
			util.Log("Error adding result", dbErr)
			return
		}
		if compErr != nil{
			continue
		}
		runRes := tool.RunTest(f.Id, ti, test, dir)
		dbErr = addResult(runRes)
		if dbErr != nil{
			util.Log("Error adding result", dbErr)
			return
		}
	}
}


/*
 Runs all available tools on a file, skipping the tools which have already been
 run.
*/
func runTools(f *submission.File, ti *tool.TargetInfo) {
	all, err := db.GetTools(bson.M{tool.LANG: ti.Lang})
	if err != nil {
		util.Log("Error retrieving tools: ", ti.Lang)
		return
	}
	for _, t := range all {
		//Check if tool has already been run
		if _, ok := f.Results[t.Name];ok {
			continue
		}
		res := tool.RunTool(f.Id, ti, t, map[string]string{})
		err = addResult(res)
		if err != nil {
			util.Log("Could not add result: ", err)
			return
		}
	}
}


/*
Extracts files from archive and processes them.
*/
func processArchive(archive *submission.File) {
	files, err := util.UnzipToMap(archive.Data)
	if err != nil {
		util.Log("Bad archive: ", err)
		return
	}
	for name, data := range files {
		info, err := submission.ParseName(name)
		if err != nil {
			util.Log("Error reading file metadata: ", name, err)
			continue
		}
		matcher := bson.M{submission.INFO: info}
		f, err := db.GetFile(matcher)
		if err != nil {
			f = submission.NewFile(archive.SubId, info, data)
			err = db.AddFile(f)
			if err != nil {
				util.Log("Error storing file: ", f.Id, err)
				continue
			}
		}
		processFile(f)
	}
}


type compFunc func(fileId bson.ObjectId, ti *tool.TargetInfo) bool

/*
Compiles a java source file and writes the results thereof to the database.
*/
func compile(fileId bson.ObjectId, ti *tool.TargetInfo) bool {
	comp, err := db.GetTool(ti.GetCompiler())
	if err != nil {
		util.Log(ti.Lang+" compiler not found: ", err)
		return false
	}
	res := tool.RunTool(fileId, ti, comp, map[string]string{tool.CP : ti.Dir})
	err = addResult(res)
	if err != nil {
		util.Log("Could not add result:", res, err)
		return false
	}
	return true
}

/*
Compiles a java source file and writes the results thereof to the database.
*/
func alreadyCompiled(fileId bson.ObjectId, ti *tool.TargetInfo)bool {
	comp, err := db.GetTool(ti.GetCompiler())
	if err != nil {
		util.Log(ti.Lang+" compiler not found: ", err)
		return false
	}
	res := tool.NewResult(fileId, comp.Id, comp.Name, comp.OutName, comp.ErrName, []byte(""), []byte(""))
	err = addResult(res)
	if err != nil {
		util.Log("Could not add result:", res, err)
		return false
	}
	return true	
}


func addResult(res *tool.Result)error {
	matcher := bson.M{submission.ID: res.FileId}
	change := bson.M{db.SET: bson.M{submission.RES+"." + res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}

/*Tool configuration for testing
func init() {
	_, err := db.GetTool(bson.M{NAME: "findbugs"})
	if err != nil {
		fb := &Tool{bson.NewObjectId(), "findbugs", "java", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", "warnings", []string{"java", "-jar"}, []string{"-textui", "-low"}, bson.M{"":""}, PKG_PATH}
		db.AddTool(fb)
	}
	_, err = db.GetTool(bson.M{NAME: COMPILE})
	if err != nil {
		javac := &Tool{bson.NewObjectId(), COMPILE, "java", "javac", "warnings", "errors", []string{""}, []string{"-implicit:class"}, bson.M{CP:""}, FILE_PATH}
		db.AddTool(javac)
	}
}*/

