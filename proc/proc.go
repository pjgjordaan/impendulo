package proc

import (
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/sub"
	"github.com/godfried/cabanga/tools"
	"github.com/godfried/cabanga/utils"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
)

const (
	INCOMPLETE = "incomplete.gob"
)


var builder *testBuilder

type testBuilder struct {
	tests   map[string]bool
	m       *sync.Mutex
	testDir string
}

type Item struct {
	FileId bson.ObjectId
	SubId  bson.ObjectId
}

type status struct {
	busy chan  bson.ObjectId
	done chan  bson.ObjectId
}

func newTestBuilder() *testBuilder {
	dir := filepath.Join(os.TempDir(), "tests")
	return &testBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

/*
Extracts project tests from db to filesystem for execution.
*/
func (this *testBuilder) setup(project string) (ret bool) {
	this.m.Lock()
	ret = true
	if !this.tests[project] {
		tests, ok := getTests(project)
		if !ok {
			ret = false
		} else {
			err := utils.Unzip(filepath.Join(this.testDir, project), tests.Data)
			if err != nil {
				utils.Log("Error extracting tests:", project, err)
				ret = false
			} else {
				this.tests[project] = true
			}
		}
	}
	this.m.Unlock()
	return ret
}

/*
Retrieves project tests from database.
*/
func getTests(project string) (*sub.File, bool) {
	s, ok := LoadSubmission(db.GetOne, bson.M{sub.PROJECT: project, sub.MODE: sub.TEST_MODE})
	if !ok {
		return nil, false
	}
	return LoadFile(db.GetOne, bson.M{"subid": s.Id})
}

/*
Spawns new processing routines for each file which needs to be processed.
*/
func Serve(items chan Item) {
	builder = newTestBuilder()
	// Start handlers
	stat := new(status)
	go statusListener(stat)
	go func() {
		stored := getStored()
		for subId, _ := range stored {
			processStored(subId, stat)
		}
	}()
	subs := make(map[bson.ObjectId]chan bson.ObjectId)
	for item := range items {
		if ch, ok := subs[item.SubId]; ok {
			ch <- item.FileId
		} else {
			subs[item.SubId] = make(chan bson.ObjectId)
			go processNew(item.SubId, subs[item.SubId], stat)
			subs[item.SubId] <- item.FileId
		}
	}
}

/*
Retrieves incompletely processed submissions from  the filesystem.
*/
func getStored()  map[bson.ObjectId]bool {
	stored, err := utils.LoadMap(INCOMPLETE)
	if err != nil{
		utils.Log("Unable to read stored map: ", err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}


/*
Retrieves incompletely processed submissions from  the filesystem.
*/
func saveActive(active map[bson.ObjectId]bool) {
	err := utils.SaveMap(active, INCOMPLETE)
	if err != nil{
		utils.Log("Unable to save active processes map: ", err)
	}
}

/*
Listens for new submissions and adds them to the map of active processes. Listens for completed submissions and 
removes them from the active process map. Listens for Kill or Interrupt signals and saves the active processes if
these signals are detected. 
*/
func statusListener(stat *status) {
	active := getStored()
	quitCh := make(chan os.Signal)
	signal.Notify(quitCh, os.Kill, os.Interrupt)
	for {
		select {
		case id := <-stat.busy:
			active[id] = true
		case id := <-stat.done:
			delete(active, id)
		case sig := <-quitCh:
			utils.Log("Received interrupt signal: ", sig)
			saveActive(active)
			os.RemoveAll(filepath.Join(os.TempDir(), "tests"))
			os.Exit(0)
		}
	}
}

/*
Processes an incomplete submission. Retrieves all files in the submission from the db
and processes them. 
*/
func processStored(subId bson.ObjectId, stat *status) {
	stat.busy <- subId
	count := 0
	for {
		matcher := bson.M{sub.SUBID: subId, sub.NUM: count}
		fmap, err := db.GetOne(db.FILES, matcher)
		if err != nil {
			break
		}
		f, err := sub.ReadFile(fmap)
		if err != nil {
			utils.Log("Error reading file:", err)
			break
		}
		processFile(f)
	}
	err := os.RemoveAll(filepath.Join(os.TempDir(), subId.Hex()))
	if err != nil {
		utils.Log("Error cleaning files:", err)
	}
	stat.done <- subId
}


/*
Processes a new submission. Listens for incoming files and processes them.
*/
func processNew(subId bson.ObjectId, fileIds chan bson.ObjectId, stat *status) {
	stat.busy <- subId
	for fileId := range fileIds {
		processId(fileId)
	}
	err := os.RemoveAll(filepath.Join(os.TempDir(), subId.Hex()))
	if err != nil {
		utils.Log("Error cleaning files:", err)
	}
	stat.done <- subId
}

/*
Loads a file from the db and processes it.
*/
func processId(fId bson.ObjectId) {
	f, ok := LoadFile(db.GetById, fId)
	if !ok {
		return
	}
	processFile(f)
}

/*
Processes a file according to its type.
*/
func processFile(f *sub.File) {
	utils.Log("Processing file: ", f.Id)
	t := f.Type()
	if t == sub.ARCHIVE {
		processArchive(f)
		db.RemoveById(db.FILES, f.Id)
	} else if t == sub.SRC {
		evaluate(f, false)
	} else if t == sub.EXEC {
		evaluate(f, true)
	}
	utils.Log("Processed file: ", f.Id)
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
		f, compiled = LoadFile(db.GetById, f.Id)
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
}

/*
Loads a file from the db and returns it.
*/
func LoadFile(getter db.SingleGet, matcher interface{}) (*sub.File, bool) {
	fmap, err := getter(db.FILES, matcher)
	if err != nil {
		utils.Log("Error reading from db:", err)
		return nil, false
	}
	f, err := sub.ReadFile(fmap)
	if err != nil {
		utils.Log("Error reading file:", err)
		return nil, false
	}
	return f, true
}

/*
Loads a submission from the db and returns it.
*/
func LoadSubmission(getter db.SingleGet, matcher interface{}) (*sub.Submission, bool) {
	smap, err := getter(db.SUBMISSIONS, matcher)
	if err != nil {
		utils.Log("Error reading from db:", err)
		return nil, false
	}
	s, err := sub.ReadSubmission(smap)
	if err != nil {
		utils.Log("Error reading submission:", err)
		return nil, false
	}
	return s, true

}

/*
 Saves file to filesystem. Returns file info used by tools & tests.
*/
func setupFile(f *sub.File) (*tools.TargetInfo, bool) {
	s, ok := LoadSubmission(db.GetById, f.SubId)
	if !ok {
		return nil, false
	}
	dir := filepath.Join(os.TempDir(), f.SubId.Hex())
	ti := tools.NewTarget(s.Project, f.InfoStr(sub.NAME), s.Lang, f.InfoStr(sub.PKG), dir)
	err := utils.SaveFile(filepath.Join(dir, ti.Package), ti.FullName(), f.Data)
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
func processArchive(archive *sub.File) {
	files, err := utils.UnZip(archive.Data)
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
		f, ok := LoadFile(db.GetOne, bson.M{sub.INFO: info})
		if !ok {
			f = sub.NewFile(archive.SubId, info, data)
			err = db.AddOne(db.FILES, f)
			if err != nil {
				utils.Log("Error storing file: ", f.Id, err)
				continue
			}
		}
		processFile(f)
	}
}
