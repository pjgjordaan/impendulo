package proc

import (
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/tools"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"sync"
	"os/signal"
	"encoding/gob"
	"fmt"
)

type TestBuilder struct {
	Tests   map[string]bool
	m       *sync.Mutex
	TestDir string
}

func NewTestBuilder() *TestBuilder {
	dir := filepath.Join(os.TempDir(), "tests")
	return &TestBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

/*
Extracts project tests from db to filesystem for execution.
*/
func (t *TestBuilder) Setup(project string)(ret bool){
	t.m.Lock()
	if !t.Tests[project] {
		tests, err := GetTests(project)
		if err == nil{
			err := utils.Unzip(filepath.Join(t.TestDir, project), tests.Data)
			if err != nil {
				panic(err)
			}
			t.Tests[project] = true
			ret = true
		} 
	}
	t.m.Unlock()
	return ret
}

/*
Retrieves project tests from database.
*/
func GetTests(project string) (tests *sub.File, err error) {
	smap, err := db.GetOne(db.SUBMISSIONS, bson.M{sub.PROJECT: project, sub.MODE: sub.TEST})
	if err != nil {
		return tests, err
	}
	s := sub.ReadSubmission(smap)
	fmap, err := db.GetOne(db.FILES, bson.M{"subid": s.Id})
	if err != nil {
		return tests, err
	}
	tests = sub.ReadFile(fmap)
	return tests, err
}

var testBuilder *TestBuilder

/*
Spawns new processing routines for each file which needs to be processed.
*/
func Serve(files chan *sub.File) {
	testBuilder = NewTestBuilder()
	// Start handlers
	status := make(chan *Status)
	go statusListener(status, getQueued())
	for f := range files {
		go Process(f, status)
	}
}

type Status struct{
	Id bson.ObjectId
	Phase int
}

const(
	BUSY = iota
DONE 
)

func getQueued()(queued map[bson.ObjectId] *Status){
	f, err := os.Open(filepath.Join(utils.BASE_DIR, "queue.gob"))
	if err == nil{
		dec := gob.NewDecoder(f) 
		err = dec.Decode(&queued)
	}
	if err != nil{
		utils.Log("Unable to read queue: ", err)
		queued = make(map[bson.ObjectId] *Status)
	}
	fmt.Println("Reading: ",queued)
	return queued
}



func saveQueued(queued map[bson.ObjectId] *Status) error{
	f, err := os.Create(filepath.Join(utils.BASE_DIR, "queue.gob"))
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	fmt.Println("Saving: ",queued)
	return enc.Encode(&queued)
}


/*
Keeps track of currently active processing routines.
*/
func statusListener(status chan *Status, active map[bson.ObjectId] *Status){
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)
	for {
		select{
		case stat := <- status:
			if stat.Phase == DONE{
				delete(active, stat.Id)
			} else if _, ok := active[stat.Id]; !ok{
				active[stat.Id] = stat
			}
		case sig := <- signals:
			utils.Log("Received interrupt signal: ", sig)
			fmt.Println("Received interrupt signal: ", sig)
			err := saveQueued(active)
			if err != nil{
				utils.Log("Saving queue failed: ", err)
			}
			fmt.Println("Saved queue: ", err)
			os.Exit(0)
		}
	}
}


/*
Adds a file to the database and processes it according to its type.
*/
func Process(f *sub.File, status chan *Status) {
	t := f.Type()
	if t == sub.ARCHIVE {
		ProcessArchive(f, status)
	} else {
		err := db.AddOne(db.FILES, f)
		if err != nil {
			panic(err)
		}
		fmt.Println(t)
		status <- &Status{f.Id, BUSY}
		if t == sub.SRC {
			ProcessSource(f)
		} else if t == sub.EXEC {
			ProcessExec(f)
		} else if t == sub.CHANGE {
			ProcessChange(f)
		} else if t == sub.TEST {
			ProcessTest(f)
		}
		status <- &Status{f.Id, DONE}
	}
}

/*
Compiles a source file. Runs tools and tests on compiled file.
*/
func ProcessSource(src *sub.File) {
	smap, err := db.GetById(db.SUBMISSIONS, src.SubId)
	if err != nil {
		utils.Log("Submission not found:", err)
		return
	}
	s := sub.ReadSubmission(smap)
	ti := setupSource(src, s)
	if tools.Compile(src.Id, ti) {
		runTests(src, s.Project, ti)
		tools.RunTools(src.Id, ti)
	} else{
		fmt.Println("Could not compile: ", src)
	}
	//clean up
	err = os.RemoveAll(ti.Dir)
	if err != nil {
		panic(err)
	}
}
/*
Saves source file to filesystem.
*/
func setupSource(src *sub.File, s *sub.Submission) (ti *tools.TargetInfo) {
	dir := filepath.Join(os.TempDir(), src.Id.Hex())
	ti = tools.NewTarget(src.InfoStr(sub.NAME),s.Lang, src.InfoStr(sub.PKG), dir)
	err := utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), src.Data)
	if err != nil {
		panic(err)
	}
	return ti
}

/*
Sets up and runs test on executable file.
*/
func runTests(src *sub.File, project string, ti *tools.TargetInfo) {
	if testBuilder.Setup(project){
		tools.RunTests(src.Id, project, ti)
	} else {
		utils.Log("No tests found for ", project)
	}
}

func ProcessExec(s *sub.File) {
}

func ProcessChange(s *sub.File) {
}

func ProcessTest(s *sub.File) {
}

/*
Extracts files from archive and processes them.
*/
func ProcessArchive(s *sub.File, status chan *Status) {
	files, err := utils.ReadZip(s.Data)
	if err != nil {
		utils.Log("Bad archive: ", err)
		return
	}
	for name, data := range files {
		info := sub.ParseName(name)
		f := sub.NewFile(s.SubId, info, data)
		Process(f, status)
	}
}
