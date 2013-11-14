package receiver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"net"
	"strconv"
	"testing"
	"time"
)

type (
	clientSpawner struct {
		users           map[string]string
		mode            string
		numFiles, rport int
		data            []byte
	}
	client struct {
		uname, pword, mode string
		projectId          bson.ObjectId
		submission         *project.Submission
		conn               net.Conn
	}
)

var (
	pport = 9000
)

func init() {
	util.SetErrorLogging("a")
	util.SetInfoLogging("a")
	go func() {
		processing.Serve(10)
	}()
}

func (this *clientSpawner) spawn() (*client, bool) {
	for uname, pword := range this.users {
		c := &client{
			uname: uname,
			pword: pword,
			mode:  this.mode,
		}
		delete(this.users, uname)
		return c, true
	}
	return nil, false
}

func addData(numUsers int) (users map[string]string, err error) {
	p := project.New("Triangle", "user", "Java", []byte{})
	err = db.Add(db.PROJECTS, p)
	if err != nil {
		return
	}
	users = make(map[string]string, numUsers)
	for i := 0; i < numUsers; i++ {
		uname := "user" + strconv.Itoa(i)
		users[uname] = "password"
		err = db.Add(db.USERS, user.New(uname, "password"))
		if err != nil {
			return
		}
	}
	return
}

func receive(port int) {
	started := make(chan struct{})
	go func() {
		started <- struct{}{}
		Run(port, &SubmissionSpawner{})
	}()
	<-started
}

func (this *client) login(port int) (projectId bson.ObjectId, err error) {
	this.conn, err = net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	req := map[string]interface{}{
		REQ:          LOGIN,
		db.USER:      this.uname,
		db.PWORD:     this.pword,
		project.MODE: this.mode,
	}
	err = write(this.conn, req)
	if err != nil {
		return
	}
	read, err := util.ReadData(this.conn)
	if err != nil {
		return
	}
	var infos []*ProjectInfo
	err = json.Unmarshal(read, &infos)
	if err != nil {
		return
	}
	if len(infos) == 0 {
		err = errors.New("No projects found.")
	} else {
		projectId = infos[0].Project.Id
	}
	return
}

func (this *client) create(projectId bson.ObjectId) (err error) {
	req := map[string]interface{}{
		REQ:          NEW,
		db.PROJECTID: projectId,
		db.TIME:      util.CurMilis(),
	}
	err = write(this.conn, req)
	if err != nil {
		return
	}
	read, err := util.ReadData(this.conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(read, &this.submission)
	return
}

func (this *client) logout() (err error) {
	req := map[string]interface{}{
		REQ: LOGOUT,
	}
	err = write(this.conn, req)
	if err != nil {
		return
	}
	err = readOk(this.conn)
	return
}

func (this *client) send(numFiles int, data []byte) (err error) {
	switch this.mode {
	case project.FILE_MODE:
		err = this.sendFile(numFiles, data)
	case project.ARCHIVE_MODE:
		err = this.sendArchive(data)
	default:
		err = fmt.Errorf("Unsupported mode %s.", this.mode)
	}
	return
}

func (this *client) sendArchive(data []byte) (err error) {
	req := map[string]interface{}{
		REQ: SEND,
	}
	err = write(this.conn, req)
	if err != nil {
		return
	}
	err = readOk(this.conn)
	if err != nil {
		return
	}
	_, err = this.conn.Write(data)
	if err != nil {
		return
	}
	_, err = this.conn.Write([]byte(util.EOT))
	if err != nil {
		return
	}
	err = readOk(this.conn)
	return
}

func (this *client) sendFile(numFiles int, data []byte) (err error) {
	req := map[string]interface{}{
		REQ:          SEND,
		project.TYPE: project.SRC,
		db.NAME:      "Triangle.java",
		db.PKG:       "triangle",
	}
	for i := 0; i < numFiles; i++ {
		req[db.TIME] = util.CurMilis()
		err = write(this.conn, req)
		if err != nil {
			return
		}
		err = readOk(this.conn)
		if err != nil {
			return
		}
		_, err = this.conn.Write(data)
		if err != nil {
			return
		}
		_, err = this.conn.Write([]byte(util.EOT))
		if err != nil {
			return
		}
		err = readOk(this.conn)
		if err != nil {
			return
		}
	}
	return
}

func write(conn net.Conn, data interface{}) (err error) {
	err = util.WriteJson(conn, data)
	if err != nil {
		return
	}
	_, err = conn.Write([]byte(util.EOT))
	return
}

func readOk(conn net.Conn) (err error) {
	read, err := util.ReadData(conn)
	if err == nil && !bytes.HasPrefix(read, []byte(OK)) {
		err = fmt.Errorf("Unexpected reply %s.", string(read))
	}
	return
}

func loadZip(fileNum int) ([]byte, error) {
	data := make(map[string][]byte)
	start := 1377870393875
	for i := 0; i < fileNum; i++ {
		name := fmt.Sprintf("triangle_Triangle.java_%d_c", start+i)
		data[name] = fileData
	}
	return util.ZipMap(data)
}

func testReceive(spawner *clientSpawner) (err error) {
	errChan := make(chan error)
	nU := len(spawner.users)
	ok := true
	cli, ok := spawner.spawn()
	for ok {
		go func(c *client) {
			var err error
			defer func() {
				errChan <- err
			}()
			projectId, err := c.login(spawner.rport)
			if err != nil {
				return
			}
			err = c.create(projectId)
			if err != nil {
				return
			}
			err = c.send(spawner.numFiles, spawner.data)
			if err != nil {
				return
			}
			err = c.logout()
			return
		}(cli)
		cli, ok = spawner.spawn()
	}
	done := 1
	for done < nU && err == nil {
		err = <-errChan
		done++
	}
	time.Sleep(100 * time.Millisecond)
	ierr := processing.WaitIdle()
	if err == nil && ierr != nil {
		err = ierr
	}
	return
}

func TestFile(t *testing.T) {
	nU, nF, rport := 1, 2, 8000
	db.Setup(db.TEST_CONN + "tf")
	db.DeleteDB(db.TEST_DB + "tf")
	db.Setup(db.TEST_CONN + "tf")
	defer db.DeleteDB(db.TEST_DB + "tf")
	users, err := addData(nU)
	if err != nil {
		t.Error(err)
	}
	receive(rport)
	fspawner := &clientSpawner{
		mode:     project.FILE_MODE,
		data:     fileData,
		users:    users,
		numFiles: nF,
		rport:    rport,
	}
	err = testReceive(fspawner)
	if err != nil {
		t.Error(err)
	}
}

func TestArchive(t *testing.T) {
	nU, nF, rport := 1, 2, 8010
	db.Setup(db.TEST_CONN + "ta")
	db.DeleteDB(db.TEST_DB + "ta")
	db.Setup(db.TEST_CONN + "ta")
	defer db.DeleteDB(db.TEST_DB + "ta")
	users, err := addData(nU)
	if err != nil {
		t.Error(err)
	}
	zipData, err := loadZip(nF)
	if err != nil {
		t.Error(err)
	}
	receive(rport)
	aspawner := &clientSpawner{
		mode:     project.ARCHIVE_MODE,
		data:     zipData,
		users:    users,
		numFiles: nF,
		rport:    rport,
	}
	err = testReceive(aspawner)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkFile(b *testing.B) {
	nU, nF, rport := 2, 1, 8020
	db.Setup(db.TEST_CONN + "bf")
	db.DeleteDB(db.TEST_DB + "bf")
	db.Setup(db.TEST_CONN + "bf")
	defer db.DeleteDB(db.TEST_DB + "bf")
	users, err := addData(nU)
	if err != nil {
		b.Error(err)
	}
	fspawner := &clientSpawner{
		mode:     project.FILE_MODE,
		data:     fileData,
		users:    users,
		numFiles: nF,
		rport:    rport,
	}
	receive(rport)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = testReceive(fspawner)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkArchive(b *testing.B) {
	nU, nF, rport := 2, 1, 8030
	db.Setup(db.TEST_CONN + "ba")
	db.DeleteDB(db.TEST_DB + "ba")
	db.Setup(db.TEST_CONN + "ba")
	defer db.DeleteDB(db.TEST_DB + "ba")
	users, err := addData(nU)
	if err != nil {
		b.Error(err)
	}
	zipData, err := loadZip(nF)
	if err != nil {
		b.Error(err)
	}
	aspawner := &clientSpawner{
		mode:     project.ARCHIVE_MODE,
		data:     zipData,
		users:    users,
		numFiles: nF,
		rport:    rport,
	}
	receive(rport)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = testReceive(aspawner)
		if err != nil {
			b.Error(err)
		}
	}
}

var fileData = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)
