package processing

import (
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/tool"
	"github.com/godfried/cabanga/tool/java"
	"github.com/godfried/cabanga/tool/findbugs"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
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
	go Monitor(INCOMPLETE, busy, done, quit)
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

//ProcessStored processes an incompletely processed submission. 
//It retrieves files in the submission from the db and sends them on fileChan to be processed. 
func ProcessStored(subId bson.ObjectId, subChan chan *submission.Submission, fileChan chan *submission.File) {
	sub, err := db.GetSubmission(bson.M{submission.ID: subId})
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
		matcher := bson.M{submission.SUBID: subId, submission.INFO+"."+submission.NUM: count}
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
	util.Log("Processing submission", sub)
	dir := filepath.Join(os.TempDir(), sub.Id.Hex())
	//defer os.RemoveAll(dir)
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
	util.Log("Processed submission", sub)
	done <- sub.Id
}


//ProcessFile processes a file according to its type. 
func ProcessFile(f *submission.File, dir string, test *TestRunner) error {
	util.Log("Processing file", f.Id)
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
	util.Log("Processed file", f.Id)
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
		err = test.Execute(f, target.Dir)
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
	var res *tool.Result
	var err error
	javac := java.NewJavac(ti.Dir)
	if isSource{
		res, err = javac.Run(fileId, ti) 
		if err != nil {
			return false, err
		}
	} else{
		res = tool.NewResult(fileId, javac, []byte(""), []byte(""), nil)
	} 
	util.Log("Compile result", res)
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

//RunTools runs all available tools on a file, skipping previously run tools.
func RunTools(f *submission.File, ti *tool.TargetInfo) error {
	fb := findbugs.NewFindBugs()
	if _, ok := f.Results[fb.GetName()]; ok {
		return nil
	}
	res, err := fb.Run(f.Id, ti)
	util.Log("Tool result", res)
	if err != nil{
		return err
	}
	return AddResult(res)
}
