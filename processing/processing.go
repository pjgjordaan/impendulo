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
	"strings"
)

const INCOMPLETE = "incomplete.gob"

//Serve spawns new processing routines for each new submission received on subChan.
//New files are received on fileChan and then sent to the relevant submission process.
//Incomplete submissions are read from disk and reprocessed using the ProcessStored function.   
func Serve(subChan chan *submission.Submission, fileChan chan *submission.File) {
	// Start handlers
	busy := make(chan bson.ObjectId)
	done := make(chan bson.ObjectId)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Kill, os.Interrupt)
	go StatusListener(INCOMPLETE, busy, done, quit)
	go func(){
		stored := getStored(INCOMPLETE)
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
			if ch, ok := subs[sub.Id]; ok {
				close(ch)
				delete(subs, sub.Id)
			} else{
				subs[sub.Id] = make(chan *submission.File)
				go ProcessSubmission(sub, subs[sub.Id], busy, done)
			} 
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
func getStored(fname string) map[bson.ObjectId]bool {
	stored, err := util.LoadMap(filepath.Join(util.BaseDir(), fname))
	if err != nil {
		fmt.Println(err)
		util.Log(err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}

//saveActive saves active submissions to the filesystem.
func saveActive(fname string, active map[bson.ObjectId]bool)error {
	err := util.SaveMap(active, filepath.Join(util.BaseDir(), fname))
	if err != nil {
		return err
	}
	return nil
}

//StatusListener listens for new submissions and adds them to the map of active processes. 
//It also listens for completed submissions and removes them from the active process map.
//Finally it detects Kill and Interrupt signals, saving the active processes if they are detected.
func StatusListener(fname string, busy, done chan bson.ObjectId, quit chan os.Signal) {
	active := getStored(fname)
	for {
		select {
		case id := <-busy:
			active[id] = true
		case id := <-done:
			delete(active, id)
		case <-quit:
			err := saveActive(fname, active)
			if err != nil{
				util.Log(err)
			}
			//os.Exit(0)
			return
		}
	}
}

//ProcessStored processes an incompletely processed submission. 
//It retrieves files in the submission from the db and sends them on fileChan to be processed. 
func ProcessStored(subId bson.ObjectId, subChan chan *submission.Submission, fileChan chan *submission.File) {
	sub, err := db.GetSubmission(bson.M{submission.SUBID: subId})
	if err != nil {
		util.Log(err)
		return
	}
	total, err := db.Count(db.FILES, bson.M{submission.SUBID: subId})
	if err != nil {
		util.Log(err)
		return
	}
	subChan <- sub
	count := 0
	for count < total{
		matcher := bson.M{submission.SUBID: subId, submission.NUM: count}
		file, err := db.GetFile(matcher)
		if err != nil {
			util.Log(err)
			return
		}
		fileChan <- file
		count ++
	}
	subChan <- sub
}

//ProcessSubmission processes a new submission.
//It listens for incoming files on fileChan and processes them.
func ProcessSubmission(sub *submission.Submission, fileChan chan *submission.File, busy, done chan bson.ObjectId) {
	busy <- sub.Id
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	defer os.RemoveAll(dir)
	test := SetupTests(sub.Project, sub.Lang,  dir)
	for {
		file, ok := <- fileChan
		if !ok{
			break
		}
		err := ProcessFile(file, dir, test)
		if err != nil {
			util.Log(err)
			return
		}
	}
	done <- sub.Id
}


//ProcessFile processes a file according to its type. 
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

//ProcessArchive extracts files from an archive and processes them.
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

//Evaluate evaluates a source or compiled file by attempting to run tests and tools on it.
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

//ExtractFile saves a file to filesystem.
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


//Compile compiles a java source file and saves the results thereof.
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

//SetupTests extracts a project's tests from db to filesystem for execution.
//It creates and returns a new TestRunner.
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


//TestRunner is used to run tests on files compiled files.
type TestRunner struct {
	Project string
	Names []string
	Lang    string
	Dir     string
}

//Execute sets up and runs tests on a compiled file. 
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

const(
	JUNIT_EXEC = "org.junit.runner.JUnitCore"
	JUNIT_JAR = "/usr/share/java/junit4.jar"
)

//Compile compiles a test for the current file. 
func (this *TestRunner) Compile(testName string, f *submission.File, target *tool.TargetInfo) (bool, error) {
	if _, ok := f.Results[testName+"_compile"]; ok {
		return true, nil
	}
	cp := this.Dir+":"+JUNIT_JAR
	stderr, stdout, ok, err := tool.RunCommand("javac", "-cp", cp, "-implicit:class", filepath.Join(this.Dir,"testing",testName))
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

//Run runs a test on the current file.
func (this *TestRunner) Run(testName string, f *submission.File, target *tool.TargetInfo) error {
	if _, ok := f.Results[testName+"_run"]; ok {
		return nil
	}
	cp := this.Dir+":"+JUNIT_JAR
	env := "-Ddata.location="+this.Dir
	exec := strings.Split(testName, ".")[0]
	stderr, stdout, ok, err := tool.RunCommand("java", "-cp", cp, env, JUNIT_EXEC, "testing."+exec)
	if !ok {
		return err
	}
	res := tool.NewResult(f.Id, f.Id, testName+"_run", "warnings", "errors", stdout, stderr, err)
	return AddResult(res)
}

//RunTools runs all available tools on a file, skipping previously run tools.
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
