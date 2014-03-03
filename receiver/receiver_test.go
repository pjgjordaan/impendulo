//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
		numFiles, rport uint
		files           []file
	}
	client struct {
		uname, pword, mode string
		projectId          bson.ObjectId
		submission         *project.Submission
		conn               net.Conn
	}
	file struct {
		name string
		pkg  string
		tipe project.Type
		data []byte
	}
)

func init() {
	util.SetErrorLogging("f")
	util.SetInfoLogging("f")
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

func addData(numUsers uint) (users map[string]string, err error) {
	p := project.New("Triangle", "user", "Java")
	err = db.Add(db.PROJECTS, p)
	if err != nil {
		return
	}
	users = make(map[string]string, numUsers)
	for i := 0; i < int(numUsers); i++ {
		uname := "user" + strconv.Itoa(i)
		users[uname] = "password"
		err = db.Add(db.USERS, user.New(uname, "password"))
		if err != nil {
			return
		}
	}
	return
}

func receive(port uint) {
	started := make(chan struct{})
	go func() {
		started <- struct{}{}
		Run(port, &SubmissionSpawner{})
	}()
	<-started
}

func (this *client) login(port uint) (projectId bson.ObjectId, err error) {
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

func (this *client) send(numFiles uint, files []file) (err error) {
	switch this.mode {
	case project.FILE_MODE:
		err = this.sendFile(numFiles, files)
	case project.ARCHIVE_MODE:
		err = this.sendArchive(files[0])
	default:
		err = fmt.Errorf("Unsupported mode %s.", this.mode)
	}
	return
}

func (this *client) sendArchive(f file) (err error) {
	req := map[string]interface{}{
		REQ:          SEND,
		project.TYPE: f.tipe,
		db.NAME:      f.name,
		db.PKG:       f.pkg,
	}
	err = write(this.conn, req)
	if err != nil {
		return
	}
	err = readOk(this.conn)
	if err != nil {
		return
	}
	_, err = this.conn.Write(f.data)
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

func (this *client) sendFile(numFiles uint, files []file) (err error) {
	var i uint = 0
	for {
		for _, f := range files {
			if i == numFiles {
				return
			}
			req := map[string]interface{}{
				REQ:          SEND,
				project.TYPE: f.tipe,
				db.NAME:      f.name,
				db.PKG:       f.pkg,
				db.TIME:      util.CurMilis(),
			}
			err = write(this.conn, req)
			if err != nil {
				return
			}
			err = readOk(this.conn)
			if err != nil {
				return
			}
			_, err = this.conn.Write(f.data)
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
			i++
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

func loadZip(fileNum uint) ([]byte, error) {
	data := make(map[string][]byte)
	start := 1377870393875
	for i := 0; i < int(fileNum); i++ {
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
			err = c.send(spawner.numFiles, spawner.files)
			if err != nil {
				return
			}
			err = c.logout()
			return
		}(cli)
		cli, ok = spawner.spawn()
	}
	done := 0
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

func testFiles(t *testing.T, nF, nU, port uint, mode string, files []file) {
	go processing.MonitorStatus()
	go processing.Serve(processing.MAX_PROCS)
	ext := "_" + strconv.Itoa(int(port))
	db.Setup(db.TEST_CONN + ext)
	db.DeleteDB(db.TEST_DB + ext)
	db.Setup(db.TEST_CONN + ext)
	defer db.DeleteDB(db.TEST_DB + ext)
	users, err := addData(nU)
	if err != nil {
		t.Error(err)
	}
	receive(port)
	spawner := &clientSpawner{
		mode:     mode,
		files:    files,
		users:    users,
		numFiles: nF,
		rport:    port,
	}
	err = testReceive(spawner)
	if err != nil {
		t.Error(err)
	}
	err = processing.Shutdown()
	if err != nil {
		t.Error(err)
	}
}

func TestFile(t *testing.T) {
	files := []file{{"Triangle.java", "triangle", project.SRC, fileData}}
	testFiles(t, 1, 1, 8000, project.FILE_MODE, files)
	testFiles(t, 2, 3, 8000, project.FILE_MODE, files)
	files = append(files, file{"UserTests.java", "testing", project.TEST, userTestData})
	testFiles(t, 2, 1, 8000, project.FILE_MODE, files)
	testFiles(t, 4, 3, 8000, project.FILE_MODE, files)
	zipData, err := loadZip(1)
	if err != nil {
		t.Error(err)
	}
	zips := []file{{"Triangle.java", "triangle", project.ARCHIVE, zipData}}
	testFiles(t, 1, 1, 8010, project.ARCHIVE_MODE, zips)
	zipData, err = loadZip(5)
	if err != nil {
		t.Error(err)
	}
	testFiles(t, 3, 2, 8010, project.ARCHIVE_MODE, zips)
}

func benchmarkFiles(b *testing.B, nF, nU, nS, nM, port uint, mode string, files []file) {
	servers := make([]*processing.Server, nS)
	var err error
	for i := 0; i < int(nS); i++ {
		servers[i], err = processing.NewServer(processing.MAX_PROCS)
		if err != nil {
			b.Error(err)
		}
		go servers[i].Serve()
	}
	monitors := make([]*processing.Monitor, nS)
	for i := 0; i < int(nM); i++ {
		monitors[i], err = processing.NewMonitor()
		if err != nil {
			b.Error(err)
		}
		go monitors[i].Monitor()
	}
	ext := "_" + strconv.Itoa(int(port))
	db.Setup(db.TEST_CONN + ext)
	db.DeleteDB(db.TEST_DB + ext)
	db.Setup(db.TEST_CONN + ext)
	defer db.DeleteDB(db.TEST_DB + ext)
	users, err := addData(nU)
	if err != nil {
		b.Error(err)
	}
	spawner := &clientSpawner{
		mode:     mode,
		files:    files,
		users:    users,
		numFiles: nF,
		rport:    port,
	}
	receive(port)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = testReceive(spawner)
		if err != nil {
			b.Error(err)
		}
	}
	for _, s := range servers {
		err = s.Shutdown()
		if err != nil {
			b.Error(err)
		}
	}
	for _, m := range monitors {
		err = m.Shutdown()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkFile(b *testing.B) {
	files := []file{{"Triangle.java", "triangle", project.SRC, fileData}}
	benchmarkFiles(b, 2, 2, 5, 5, 8020, project.FILE_MODE, files)
	files = append(files, file{"UserTests.java", "testing", project.TEST, userTestData})
	benchmarkFiles(b, 4, 2, 5, 5, 8020, project.FILE_MODE, files)
	zipData, err := loadZip(2)
	if err != nil {
		b.Error(err)
	}
	zips := []file{{"Triangle.java", "triangle", project.ARCHIVE, zipData}}
	benchmarkFiles(b, 2, 2, 5, 5, 8030, project.ARCHIVE_MODE, zips)
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

var userTestData = []byte(`
package testing;

import junit.framework.TestCase;
import triangle.Triangle;

public class UserTests extends TestCase {

	public void testKselect() {
		Triangle t = new Triangle();
		int[][] values = { {6}, {6, 3}, {2, 9, 3}};
		assertEquals("Expected 21.", 21, t.maxpath(values));
	}
}
`)
