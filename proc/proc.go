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


func Javac(source string)(errBytes [] byte, err error){
	errBytes, _, err = runCommand("javac", source)
	return errBytes, err
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
	f, err := db.GetFile(r.FileId)
	if err == nil && f.IsSource(){
		params := strings.Split(f.Name, "_")
		//status := params[len(params)-1]
		//counter := strconv.Atoi(params[len(params)-2])
		//time := strconv.Atoi(params[len(params)-3])
		fname := params[len(params)-4]
		dir := filepath.Join(os.TempDir(), r.FileId.Hex(), filepath.Join(params[:len(params)-4]...))
		err = utils.SaveFile(dir, fname, f.Data)
		if err == nil{
			sub, err := db.GetSubmission(r.SubId)
			if err == nil{
				testBuilder.SetupTests(sub.Project)
			}
		}
	}
	if err != nil{
		utils.Log(err)
	}
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

func (t *TestBuilder) SetupTests(project string)(err error){
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