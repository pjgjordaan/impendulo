package proc

import(
	"labix.org/v2/mgo/bson"
	"github.com/disco-volante/intlola/utils"
	"github.com/disco-volante/intlola/db"
"strings"
"os"
"os/exec"
"io/ioutil"
"io"
"path/filepath"
"sync"
)

const MAX = 100

type Request struct{
	FileId bson.ObjectId
	SubId bson.ObjectId
}


func runCommand(args ...string)(errBytes, outBytes []byte, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	errPipe, outPipe, err := getPipes(cmd)	
	if err == nil{
		err = cmd.Start()
		if err == nil {
			errBytes, outBytes, err = getBytes(errPipe, outPipe)
			if err == nil{
				err = cmd.Wait()
			}
		}
	}
	return errBytes, outBytes, err
}


func getPipes(cmd *exec.Cmd)(errPipe, outPipe io.ReadCloser, err error){
	errPipe, err = cmd.StderrPipe()
	if err == nil{
		outPipe, err = cmd.StdoutPipe()
	}
	return errPipe, outPipe, err
}


func getBytes(errPipe, outPipe io.ReadCloser)(errBytes, outBytes []byte, err error){
	errBytes, err = ioutil.ReadAll(errPipe)
	if err != nil{
		outBytes, err = ioutil.ReadAll(outPipe)			
	}
	return errBytes, outBytes, err
}


func process(r *Request){
	src, err := setupSource(r.FileId)
	if err == nil{
		err = setupTests(r.SubId)
		if err == nil{
			RunTests(src)
		}
	}
	if err != nil{
		utils.Log(err)
	}
}

type Source struct{
	Name string
	Package string
	Ext string
	Dir string
}

func (s *Source) AbsPath() string{
	return filepath.Join(s.Dir, s.Package, s.FullName())
}


func (s *Source) FullName() string{
	return s.Name + "." + s.Ext
}

func Javac(source string, cp []string)([]byte, [] byte, error){
	return runCommand("javac", "-cp", strings.Join(cp,":"), source)
}


func RunTests(src *Source){
		//Hardcode for now
	testdir := filepath.Join(os.TempDir(), "tests")	
	classpaths := []string{src.Dir, testdir}
	tests := []*Source{&Source{"EasyTests", "testing", "java", testdir}, &Source{"AllTests", "testing", "java", testdir}}
	for _, test := range tests{
		errBytes, _, err := Javac(test.AbsPath(), classpaths)
		utils.Log(string(errBytes), err)
	}
}


func setupSource(sourceId bson.ObjectId)(src *Source, err error){
	f, err := db.GetFile(sourceId)
	if err == nil && f.IsSource(){
		//Specific to how the file names are formatted currently, should change.
		params := strings.Split(f.Name, "_")
		fname := strings.Split(params[len(params)-4], ".")
		pkg := params[len(params)-5]
		dir := filepath.Join(os.TempDir(), sourceId.Hex(), filepath.Join(params[:len(params)-5]...))
		src = &Source{fname[0], pkg, fname[1], dir} 
		err = utils.SaveFile(filepath.Join(dir, pkg), src.FullName(), f.Data)
	}
	return src, err
}

func setupTests(subId bson.ObjectId)(err error){
	sub, err := db.GetSubmission(subId)
	if err == nil{
		err = testBuilder.Setup(sub.Project)
	}
	return err
}


func handle(queue chan *Request) {
	for r := range queue {
		process(r)
	}
}

type TestBuilder struct{
	Status map[string] bool
	m *sync.Mutex
	TestDir string
}

func NewTestBuilder() *TestBuilder{
	dir := filepath.Join(os.TempDir(), "tests")
	return &TestBuilder{make(map[string]bool), new(sync.Mutex), dir}
}

func (t *TestBuilder) Setup(project string)(err error){
	t.m.Lock()
	if !t.Status[project]{
		tests, err := db.GetTests(project)
		if err == nil{
			err = utils.Unzip(t.TestDir, tests.Data)
			if err == nil{
				t.Status[project] = true
			}
		}
	} 
	t.m.Unlock()
	return err
}
var testBuilder *TestBuilder

func Serve(clientRequests chan *Request) {
	testBuilder = NewTestBuilder() 
	// Start handlers
	for i := 0; i < MAX; i++ {
		go handle(clientRequests)
	}
	utils.Log("completed")
}