package processing

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
)

const (
	INCOMPLETE = "incomplete.gob"
	SRC        = "src"
)


//Serve spawns new processing routines for each new submission received on subChan.
//New files are received on fileChan and then sent to the relevant submission process.
//Incomplete submissions are read from disk and reprocessed using the processStored function.   
func Serve(subChan chan *submission.Submission, fileChan chan *submission.File) {
	// Start handlers
	busy := make(chan bson.ObjectId)
	done := make(chan bson.ObjectId)
	go StatusListener(busy, done)
	go func(){
		stored := getStored()
		for subId, busy :=  range stored{
			if busy{
				go ProcessStored(subId, subChan, fileChan)
			}
		}
	}()
	subs := make(map[bson.ObjectId]chan *submission.File)
	for {
		select {
		case sub := <-subChan:
			subs[sub.Id] = make(chan *submission.File)
			go ProcessSubmission(sub, subs[sub.Id], busy, done)
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

//StatusListener listens for new submissions and adds them to the map of active processes. 
//It also listens for completed submissions and removes them from the active process map.
//Finally it detects Kill and Interrupt signals, saving the active processes if they are detected.
func StatusListener(busy, done chan bson.ObjectId) {
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
			os.Exit(0)
		}
	}
}

//processStored processes an incompletely processed submission. 
//It retrieves all files in the submission from the db and processes them. 
func ProcessStored(subId bson.ObjectId, subChan chan *submission.Submission, fileChan chan *submission.File) {
	sub, err := db.GetSubmission(bson.M{submission.SUBID: subId})
	if err != nil {
		util.Log(err)
		return
	}
	subChan <- sub
	count := 0
	for{
		matcher := bson.M{submission.SUBID: subId, submission.NUM: count}
		file, err := db.GetFile(matcher)
		if err != nil {
			util.Log(err)
			return
		}
		fileChan <- file
		count ++
	}				
}

//processNew processes a new submission.
//It listens for incoming files on fileChan and processes them.
func ProcessSubmission(sub *submission.Submission, fileChan chan *submission.File, busy, done chan bson.ObjectId) {
	busy <- sub.Id
	dir := filepath.Join(os.TempDir(), sub.Id.Hex(), SRC)
	defer os.RemoveAll(dir)
	test := SetupTests(sub.Project, sub.Lang,  dir)
	for file := range fileChan {
		err := ProcessFile(file, dir, test)
		if err != nil {
			util.Log(err)
			return
		}
	}
	done <- sub.Id
}


//processFile processes a file according to its type. 
//If an error occurs, it is returned.
func ProcessFile(f *submission.File, dir string, test *TestRunner) error {
	t := f.Type()
	if t == submission.ARCHIVE {
		err := ProcessArchive(f, dir, test)
		if err != nil {
			return err
		}
		db.RemoveFileByID(f.Id)
	} else if t == submission.SRC || t == submission.EXEC {
		err := Evaluate(f, dir, test, t == submission.SRC)
		if err != nil {
			return err
		}
	}
	return nil
}

//processArchive extracts files from an archive and processes them.
func ProcessArchive(archive *submission.File, dir string, test *TestRunner) error {
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
		err = ProcessFile(f, dir, test)
		if err != nil {
			return err
		}
	}
	return nil
}

//evaluate evaluates a source or compiled file by attempting to run tests and tools on it.
func Evaluate(f *submission.File, dir string, test *TestRunner, isSource bool) error {
	target, err := ExtractFile(f, dir)
	if err != nil {
		return err
	}
	compiled, err := Compile(f.Id, target, isSource)
	if err != nil {
		return err
	}
	if !compiled {
		return nil
	}
	f, err = db.GetFile(bson.M{submission.ID: f.Id})
	if err != nil {
		return err
	}
	if test != nil {
		err = test.Execute(f, target)
		if err != nil {
			return err
		}
	}
	err = RunTools(f, target)
	if err != nil {
		return err
	}
	return nil
}

//extractFile saves a file to filesystem.
//It returns file info used by tools & tests.
func ExtractFile(f *submission.File, dir string) (*tool.TargetInfo, error) {
	matcher := bson.M{submission.ID: f.SubId}
	s, err := db.GetSubmission(matcher)
	if err != nil {
		return nil, err
	}
	ti := tool.NewTarget(s.Project, f.InfoStr(submission.NAME), s.Lang, f.InfoStr(submission.PKG), dir)
	err = util.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), f.Data)
	if err != nil {
		return nil, err
	}
	return ti, nil
}


//compile compiles a java source file and writes the results thereof to the database.
//It returns true if compiled successfully.
func Compile(fileId bson.ObjectId, ti *tool.TargetInfo, isSource bool) (bool, error) {
	matcher := bson.M{submission.LANG: submission.LANG, submission.NAME: "compile"}
	compiler, err := db.GetTool(matcher)
	if err != nil {
		return false, err
	}
	var res *tool.Result
	if isSource{
		res, err := compiler.Run(fileId, ti, map[string]string{"-cp": ti.Dir}) 
		if err != nil {
			return false, err
		}
		err = AddResult(res)
		if err != nil {
			return false, err
		}
	}else{
		res = tool.ToolResult(fileId, compiler, []byte(""), []byte(""), nil)
	} 
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}



//AddResult adds a tool result to the db.
//It updates the associated file's list of results to point to this new result.
func AddResult(res *tool.Result) error {
	matcher := bson.M{submission.ID: res.FileId}
	change := bson.M{db.SET: bson.M{submission.RES + "." + res.Name: res.Id}}
	err := db.Update(db.FILES, matcher, change)
	if err != nil {
		return err
	}
	return db.AddResult(res)
}



//setup extracts a project's tests from db to filesystem for execution.
//It returns true if this was successful.
func SetupTests(project, lang, dir string) *TestRunner {
	testMatcher := bson.M{submission.PROJECT: project, submission.LANG: lang}
	test, err := db.GetTest(testMatcher)
	if err != nil {
		util.Log(err)
		return nil
	}
	err = util.Unzip(dir, test.Tests)
	if err != nil {
		util.Log(err)
		return nil
	}
	err = util.Unzip(dir, test.Data)
	if err != nil {
		util.Log(err)
		return nil
	}
	return &TestRunner{test.Project, test.Names, test.Lang, dir}
}


//TestInfo stores information about test files.
type TestRunner struct {
	Project string
	//File name without extension
	Names []string
	//Language file is written in
	Lang    string
	Dir     string
}

//runTests sets up and runs tests on executable file. 
//If an error is encountered, iteration stops and the error is returned.
func (this *TestRunner)  Execute(f *submission.File, target *tool.TargetInfo) error {
	for _, name := range this.Names {
		compiled, err := this.Compile(name, f, target)
		if err != nil {
			return err
		}
		if !compiled {
			continue
		}
		err = this.Run(name, f, target)
		if err != nil {
			return err
		}
	}
	return nil
}

//compileTest
func (this *TestRunner) Compile(testName string, f *submission.File, target *tool.TargetInfo) (bool, error) {
	if _, ok := f.Results[testName+"_compile"]; ok {
		return true, nil
	}
	stderr, stdout, ok, err := tool.RunCommand("javac", "-cp", this.Dir, "-implicit:class", filepath.Join(this.Dir,"testing",testName))
	if !ok{
		return false, err
	}
	res := tool.NewResult(f.Id, f.Id, testName+"_compile", "warnings", "errors", stdout, stderr, err)
	err = AddResult(res)
	if err != nil {
		return false, err
	}
	return res.Error == nil, nil
}

//runTest
func (this *TestRunner) Run(testName string, f *submission.File, target *tool.TargetInfo) error {
	if _, ok := f.Results[testName+"_run"]; ok {
		return nil
	}
	env := "-Ddata.location="+this.Dir
	stderr, stdout, ok, err := tool.RunCommand("java", "-cp", this.Dir, env, "org.junit.runner.JUnitCore", "testing."+testName)
	if !ok {
		return err
	}
	res := tool.NewResult(f.Id, f.Id, testName+"_run", "warnings", "errors", stdout, stderr, err)
	return AddResult(res)
}

//runTools runs all available tools on a file, skipping previously run tools.
func RunTools(f *submission.File, ti *tool.TargetInfo) error {
	all, err := db.GetTools(bson.M{submission.LANG: ti.Lang})
	if err != nil {
		return err
	}
	for _, t := range all {
		if _, ok := f.Results[t.Name]; ok {
			continue
		}
		res, err := t.Run(f.Id, ti, nil)
		if err != nil {
			return err
		}
		err = AddResult(res)
		if err != nil {
			return err
		}
	}
	return nil
}



/*
//init sets up tool configuration for testing
func init() {
	db.Setup(db.DEFAULT_CONN)
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
*/
