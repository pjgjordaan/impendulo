package proc

import (
	"encoding/gob"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/tools"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
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
func (t *testBuilder) setup(project string) (ret bool) {
	t.m.Lock()
	if !t.tests[project] {
		tests, err := getTests(project)
		if err == nil {
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
func getTests(project string) (*sub.File, error) {
	smap, err := db.GetOne(db.SUBMISSIONS, bson.M{sub.PROJECT: project, sub.MODE: sub.TEST_MODE})
	if err != nil {
		return nil, err
	}
	s, err := sub.ReadSubmission(smap)
	if err != nil {
		return nil, err
	}
	fmap, err := db.GetOne(db.FILES, bson.M{"subid": s.Id})
	if err != nil {
		return nil, err
	}
	tests, err := sub.ReadFile(fmap)
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
	go func(queued map[bson.ObjectId]*status, stat chan *status) {
		for fileId, _ := range queued {
			processID(fileId, stat)
		}
	}(queued, stat)
	go statusListener(stat, queued)
	for fileId := range files {
		go processID(fileId, stat)
	}
}

/*

*/
type status struct {
	Id    bson.ObjectId
	Phase int
}

const (
	BUSY = iota
	DONE
)

func getQueued() (queued map[bson.ObjectId]*status) {
	f, err := os.Open(filepath.Join(utils.BASE_DIR, "queue.gob"))
	if err == nil {
		dec := gob.NewDecoder(f)
		err = dec.Decode(&queued)
	}
	if err != nil {
		utils.Log("Unable to read queue: ", err)
		queued = make(map[bson.ObjectId]*status)
	}
	return queued
}

func saveQueued(queued map[bson.ObjectId]*status) error {
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
func statusListener(stat chan *status, active map[bson.ObjectId]*status) {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)
	for {
		select {
		case s := <-stat:
			if s.Phase == DONE {
				delete(active, s.Id)
			} else if _, ok := active[s.Id]; !ok {
				active[s.Id] = s
			}
		case sig := <-signals:
			utils.Log("Received interrupt signal: ", sig)
			err := saveQueued(active)
			if err != nil {
				utils.Log("Saving queue failed: ", err)
			}
			os.RemoveAll(filepath.Join(os.TempDir(), "tests"))
			os.Exit(0)
		}
	}
}

func processID(fId bson.ObjectId, stat chan *status) {
	stat <- &status{fId, BUSY}
	fmap, err := db.GetById(db.FILES, fId)
	if err != nil {
		utils.Log("Error retrieving file: ", fId, err)
		return
	}
	f, err := sub.ReadFile(fmap)
	if err != nil {
		utils.Log("Error reading  file: ", fId, err)
		return
	}
	processFile(f, stat)
}

func processFile(f *sub.File, stat chan *status) {
	utils.Log("Processing file: ", f.Id)
	t := f.Type()
	if t == sub.ARCHIVE {
		processArchive(f, stat)
		db.RemoveById(db.FILES, f.Id)
	} else if t == sub.SRC {
		evaluate(f, false)
	} else if t == sub.EXEC {
		evaluate(f, true)
	}
	utils.Log("Processed file: ", f.Id)
	stat <- &status{f.Id, DONE}
}

/*
 Evaluates a submitted file (source or compiled) by 
 attempting to run tests and tools on it.
*/
func evaluate(f *sub.File, compiled bool) {
	utils.Log("Evaluating: ", f.Id, f.Info)
	ti, ok := setupFile(f)
	if !ok {
		return
	}
	if compiled {
		tools.AlreadyCompiled(f.Id, ti)
	} else {
		compiled = tools.Compile(f.Id, ti)
	}
	if compiled {
		utils.Log("Compiled: ", f.Id)
		runTests(f, ti)
		utils.Log("Tested: ", f.Id)
		tools.RunTools(f.Id, ti, f.Results)
		utils.Log("Ran tools: ", f.Id)
	} else {
		utils.Log("No compile: ", f.Id)
	}
	//clean up
	err := os.RemoveAll(ti.Dir)
	if err != nil {
		utils.Log("Error cleaning files:", err, f.Id)
	}
}

/*
 Saves file to filesystem. Returns file info used by tools & tests.  
*/
func setupFile(f *sub.File) (*tools.TargetInfo, bool) {
	smap, err := db.GetById(db.SUBMISSIONS, f.SubId)
	if err != nil {
		utils.Log("Submission not found:", err)
		return nil, false
	}
	s, err := sub.ReadSubmission(smap)
	if err != nil {
		utils.Log("Error reading submission :", err)
		return nil, false
	}
	dir := filepath.Join(os.TempDir(), f.Id.Hex())
	ti := tools.NewTarget(s.Project, f.InfoStr(sub.NAME), s.Lang, f.InfoStr(sub.PKG), dir)
	err = utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), f.Data)
	if err != nil {
		utils.Log("Error saving file: ", f.Id, err)
		return nil, false
	}
	return ti, true
}

/*
 Sets up and runs test on executable file.
*/
func runTests(src *sub.File, ti *tools.TargetInfo) {
	tests := []string{"EasyTests", "AllTests"}
	if builder.setup(ti.Project) {
		for _, test := range tests {
			//Run if not run previously
			if _, ok := src.Results[test+"_compile"]; !ok {
				tools.RunTest(src.Id, ti, test)
			}
		}
	} else {
		utils.Log("No tests found for ", ti.Project)
	}
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
		info, err := sub.ParseName(name)
		if err != nil {
			utils.Log("Error reading file metadata: ", name, err)
			continue
		}
		if _, err = db.GetOne(db.FILES, bson.M{sub.INFO: info}); err != nil {
			f := sub.NewFile(archive.SubId, info, data)
			stat <- &status{f.Id, BUSY}
			err = db.AddOne(db.FILES, f)
			if err != nil {
				utils.Log("Error storing file: ", f.Id, err)
				continue
			}
			go processFile(f, stat)
		}
	}
}
