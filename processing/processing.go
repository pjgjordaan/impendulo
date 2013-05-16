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
func Serve(subChan chan *submission.Submission, fileChan chan *submission.File) {
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
	for{
		select{
		case sub := <-subChan:
			subs[sub.Id] = make(chan *submission.File)
			go processNew(sub, subs[sub.Id], busy, done)
		case file := <- fileChan:
			if ch, ok := subs[file.SubId]; ok {
				ch <- file
			} else{
				util.Log("No channel found for submission:", file.SubId)
			}
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
	sub, err := db.GetSubmission(bson.M{submission.SUBID: subId})
	if err != nil {
			util.Log("Error retrieving submission:", err)
			return
	}
	count := 0
	hasTests := builder.setup(sub.Project)
	for {
		matcher := bson.M{submission.SUBID: subId, submission.NUM: count}
		file, err := db.GetFile(matcher)
		if err != nil {
			util.Log("Error retrieving file:", err)
			return
		}
		err = processFile(file, hasTests)
		if err != nil {
			util.Log("Error processing file:", err, file)
			return
		}
	}
	done <- subId
}


/*
Processes a new submission. Listens for incoming files and processes them.
*/
func processNew(sub *submission.Submission, fileChan chan *submission.File, busy, done chan bson.ObjectId) {
	busy <- sub.Id
	defer os.RemoveAll(filepath.Join(os.TempDir(), sub.Id.Hex()))
	hasTests := builder.setup(sub.Project)
	for file := range fileChan {
		err := processFile(file, hasTests)
		if err != nil {
			util.Log("Error processing file:", err, file)
			return
		}
	}
	done <- sub.Id
}


/*
Processes a file according to its type.
*/
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

/*
Extracts files from archive and processes them.
*/
func processArchive(archive *submission.File, hasTests bool)error {
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


/*
 Evaluates a submitted file (source or compiled) by
 attempting to run tests and tools on it.
*/
func evaluate(orig *submission.File, comp compFunc, hasTests bool)error {
	ti, err := extractFile(orig)
	if err != nil{
		return err
	}
	compiled, err := comp(orig.Id, ti)
	if err != nil{
		return err
	}
	if !compiled{
		return nil
	}
	f, err := db.GetFile(bson.M{submission.ID: orig.Id})
	if err != nil{
		return err
	}
	util.Log("Evaluating: ", f.Id, f.Info)
	if hasTests{
		err = runTests(f, ti)
		if err != nil{
			return err
		}
		util.Log("Tested: ", f.Id)
	}
	err = runTools(f, ti)
	if err != nil{
		return err
	}
	util.Log("Ran tools: ", f.Id)
	return nil
}


/*
 Saves file to filesystem. Returns file info used by tools & tests.
*/
func extractFile(f *submission.File) (*tool.TargetInfo, error) {
	matcher := bson.M{submission.ID:f.SubId}
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

/*
Compiles a java source file and writes the results thereof to the database.
Returns true if compiled successfully.
*/
func compile(fileId bson.ObjectId, ti *tool.TargetInfo)( bool, error) {
	comp, err := db.GetTool(ti.GetCompiler())
	if err != nil {
		return false, err
	}
	res := tool.RunTool(fileId, ti, comp, map[string]string{tool.CP : ti.Dir})
	err = addResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}

/*
Used by compiled files to store in the db the fact that they have successfully compiled. 
 Returns true if successful.
*/
func alreadyCompiled(fileId bson.ObjectId, ti *tool.TargetInfo)(bool, error) {
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


/*
 Sets up and runs tests on executable file. If an error is encountered, iteration stops and the error is returned.
*/
func runTests(f *submission.File, ti *tool.TargetInfo)error {
	tests := []string{"EasyTests", "AllTests"}
	for _, test := range tests {
		dir := filepath.Join(os.TempDir(), TESTS, ti.Project, SRC)
		compiled, err := compileTest(f, ti, test, dir)
		if err != nil{
			return err
		}
		if !compiled {
			continue
		}
		err = runTest(f, ti, test, dir)
		if err != nil{
			return err
		}
	}
	return nil
}


func compileTest(f *submission.File, ti *tool.TargetInfo, test, dir string)(bool, error){
	if _, ok := f.Results[test+"_"+tool.COMPILE]; ok {
		return true, nil
	}
	//Run if not run previously
	res := tool.CompileTest(f.Id, ti, test, dir)
	err := addResult(res)
	if err != nil{
		return false, err
	}
	return res.Error == nil, nil
}


func runTest(f *submission.File, ti *tool.TargetInfo, test, dir string)error{
	if _, ok := f.Results[test+"_"+tool.RUN]; ok {
		return nil
	}
	//Run if not run previously
	res := tool.RunTest(f.Id, ti, test, dir)
	return addResult(res)
}


/*
 Runs all available tools on a file, skipping the tools which have already been
 run. Returns true if there were no db errors.
*/
func runTools(f *submission.File, ti *tool.TargetInfo)error {
	all, err := db.GetTools(bson.M{tool.LANG: ti.Lang})
	if err != nil {
		return err
	}
	for _, t := range all {
		//Check if tool has already been run
		if _, ok := f.Results[t.Name];ok {
			continue
		}
		res := tool.RunTool(f.Id, ti, t, map[string]string{})
		err = addResult(res)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Adds a tool result to the db. Updates the associated file's list of results to point to this new result.
Returns any db errors.
*/
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

