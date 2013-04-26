package proc

import (
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/tools"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"path/filepath"
	"sync"
	"os/signal"
	"encoding/gob"
	"os"
)

type testBuilder struct {
	tests   map[string]bool
	m       *sync.Mutex
	testDir string
}

func newTestBuilder() *testBuilder {
	dir := filepath.Join(os.TempDir(), "tests")
	return &testBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

/*
Extracts project tests from db to filesystem for execution.
*/
func (t *testBuilder) setup(project string)(ret bool){
	t.m.Lock()
	if !t.tests[project] {
		tests, err := getTests(project)
		if err == nil{
			err := utils.Unzip(filepath.Join(t.testDir, project), tests.Data)
			if err != nil {
				utils.Log("Error extracting tests:", project, err)
				return ret 
			}
			t.tests[project] = true
			ret = true
		} 
	}
	t.m.Unlock()
	return ret
}

/*
Retrieves project tests from database.
*/
func getTests(project string) (tests *sub.File, err error) {
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

var builder *testBuilder

/*
Spawns new processing routines for each file which needs to be processed.
*/
func Serve(files chan bson.ObjectId) {
	builder = newTestBuilder()
	// Start handlers
	stat := make(chan *status)
	queued := getQueued()
	go func(queued map[bson.ObjectId] *status,stat chan *status){
		for fileId, _ := range queued{
			processID(fileId, stat)
		}
	}(queued, stat)
	go statusListener(stat, queued)
	for fileId := range files {
		go processID(fileId, stat)
	}
}

type status struct{
	Id bson.ObjectId
	Phase int
}

const(
	BUSY = iota
DONE 
)

func getQueued()(queued map[bson.ObjectId] *status){
	f, err := os.Open(filepath.Join(utils.BASE_DIR, "queue.gob"))
	if err == nil{
		dec := gob.NewDecoder(f) 
		err = dec.Decode(&queued)
	}
	if err != nil{
		utils.Log("Unable to read queue: ", err)
		queued = make(map[bson.ObjectId] *status)
	}
	return queued
}



func saveQueued(queued map[bson.ObjectId] *status) error{
	f, err := os.Create(filepath.Join(utils.BASE_DIR, "queue.gob"))
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	return enc.Encode(&queued)
}


/*
Keeps track of currently active processing routines.
*/
func statusListener(stat chan *status, active map[bson.ObjectId] *status){
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)
	for {
		select{
		case s := <- stat:
			if s.Phase == DONE{
				delete(active, s.Id)
			} else if _, ok := active[s.Id]; !ok{
				active[s.Id] = s
			}
		case sig := <- signals:
			utils.Log("Received interrupt signal: ", sig)
			err := saveQueued(active)
			if err != nil{
				utils.Log("Saving queue failed: ", err)
			}
			os.Exit(0)
		}
	}
}

func processID(fId bson.ObjectId, stat chan *status){
	stat <- &status{fId, BUSY}
	fmap, err := db.GetById(db.FILES, fId)
	if err != nil{
		utils.Log("Error retrieving file: ", fId, err)
		return
	}
	f := sub.ReadFile(fmap)
	processFile(f, stat)
}

func processFile(f *sub.File, stat chan *status){
	utils.Log("Processing file: ", f.Id)
	t := f.Type()
	if t == sub.ARCHIVE {
		processArchive(f, stat)
		db.RemoveById(db.FILES, f.Id)
	} else if t == sub.SRC {
		processSource(f)
	} else if t == sub.EXEC {
		processExec(f)
	} else if t == sub.CHANGE {
		processChange(f)
	} else if t == sub.TEST {
		processTest(f)
	}
	utils.Log("Processed file: ", f.Id)
	stat <- &status{f.Id, DONE}
}

/*
Compiles a source file. Runs tools and tests on compiled file.
*/
func processSource(src *sub.File) {
	smap, err := db.GetById(db.SUBMISSIONS, src.SubId)
	if err != nil {
		utils.Log("Submission not found:", err)
		return
	}
	s := sub.ReadSubmission(smap)
	ti, ok := setupSource(src, s)
	if !ok{
		return
	}
	if tools.Compile(src.Id, ti) {
		utils.Log("Compiled: ", src.Id)
		runTests(src, s.Project, ti)
		utils.Log("Tested: ", src.Id)
		tools.RunTools(src.Id, ti)
		utils.Log("Ran tools: ", src.Id)
	} else{
		utils.Log("No compile: ", src.Id)
	}
	//clean up
	err = os.RemoveAll(ti.Dir)
	if err != nil {
		utils.Log("Error cleaning files:", err, src.Id)
	}
}
/*
Saves source file to filesystem.
*/
func setupSource(src *sub.File, s *sub.Submission) (ti *tools.TargetInfo, ok bool) {
	dir := filepath.Join(os.TempDir(), src.Id.Hex())
	ti = tools.NewTarget(src.InfoStr(sub.NAME),s.Lang, src.InfoStr(sub.PKG), dir)
	err := utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), src.Data)
	if err != nil {
		utils.Log("Error saving file: ", src.Id, err)
		return ti, false
	}
	return ti, true
}

/*
Sets up and runs test on executable file.
*/
func runTests(src *sub.File, project string, ti *tools.TargetInfo) {
	if builder.setup(project){
		tools.RunTests(src.Id, project, ti)
	} else {
		utils.Log("No tests found for ", project)
	}
}

func processExec(s *sub.File) {
}

func processChange(s *sub.File) {
}

func processTest(s *sub.File) {
}

/*
Extracts files from archive and processes them.
*/
func processArchive(archive *sub.File, stat chan *status) {
	files, err := utils.ReadZip(archive.Data)
	if err != nil {
		utils.Log("Bad archive: ", err)
		return
	}
	for name, data := range files {
		info := sub.ParseName(name)
		if _, err = db.GetOne(db.FILES, bson.M{sub.INFO : info}); err != nil{
			f := sub.NewFile(archive.SubId, info, data)
			stat <- &status{f.Id, BUSY}
			err = db.AddOne(db.FILES, f)
			if err != nil{
				utils.Log("Error storing file: ", f.Id, err)
				return
			}
			go processFile(f, stat)
		}
	}
}
