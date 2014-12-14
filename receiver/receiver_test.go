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
	"github.com/godfried/impendulo/processor"
	"github.com/godfried/impendulo/processor/monitor"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"net"
	"strconv"
	"testing"
)

type (
	clientSpawner struct {
		users           map[string]string
		mode            string
		numFiles, rport int
		files           []file
	}
	client struct {
		uname, pword, mode string
		projectId          bson.ObjectId
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

func (c *clientSpawner) spawn() (*client, bool) {
	for u, p := range c.users {
		delete(c.users, u)
		return &client{uname: u, pword: p, mode: c.mode}, true
	}
	return nil, false
}

func addData(numUsers int) (map[string]string, error) {
	p := project.New("Triangle", "user", "Java", "A triangle.")
	if e := db.Add(db.PROJECTS, p); e != nil {
		return nil, e
	}
	if e := db.Add(db.ASSIGNMENTS, project.NewAssignment(p.Id, bson.NewObjectId(), "an assignment", "user", util.CurMilis()-10000, util.CurMilis()+100000)); e != nil {
		return nil, e
	}
	users := make(map[string]string, numUsers)
	for i := 0; i < int(numUsers); i++ {
		uname := "user" + strconv.Itoa(i)
		users[uname] = "password"
		if e := db.Add(db.USERS, user.New(uname, "password")); e != nil {
			return nil, e
		}
	}
	return users, nil
}

func receive(port int) {
	started := make(chan util.E)
	go func() {
		started <- util.E{}
		Run(port, &SubmissionSpawner{})
	}()
	<-started
}

func (c *client) login(port int) (*project.Assignment, error) {
	var e error
	c.conn, e = net.Dial("tcp", ":"+strconv.Itoa(port))
	if e != nil {
		return nil, e
	}
	if e = write(c.conn, map[string]interface{}{REQUEST: LOGIN, db.USER: c.uname, db.PWORD: c.pword, project.MODE: c.mode}); e != nil {
		return nil, e
	}
	d, e := util.ReadData(c.conn)
	if e != nil {
		return nil, e
	}
	var infos []*ProjectInfo
	if e = json.Unmarshal(d, &infos); e != nil {
		return nil, e
	}
	if len(infos) == 0 {
		return nil, errors.New("No projects found.")
	}
	return infos[0].Assignments[0].Assignment, nil
}

func (c *client) create(a *project.Assignment) error {
	if e := write(c.conn, map[string]interface{}{REQUEST: NEW, db.PROJECTID: a.ProjectId, db.ASSIGNMENTID: a.Id, db.TIME: util.CurMilis()}); e != nil {
		return e
	}
	return readOk(c.conn)
}

func (c *client) logout() error {
	if e := write(c.conn, map[string]interface{}{REQUEST: LOGOUT}); e != nil {
		return e
	}
	return readOk(c.conn)
}

func (c *client) send(numFiles int, files []file) error {
	switch c.mode {
	case project.FILE_MODE:
		return c.sendFile(numFiles, files)
	case project.ARCHIVE_MODE:
		return c.sendArchive(files[0])
	default:
		return fmt.Errorf("unsupported mode %s", c.mode)
	}
}

func (c *client) sendArchive(f file) error {
	if e := write(c.conn, map[string]interface{}{REQUEST: SEND, project.TYPE: f.tipe, db.NAME: f.name, db.PKG: f.pkg}); e != nil {
		return e
	}
	if e := readOk(c.conn); e != nil {
		return e
	}
	if _, e := c.conn.Write(f.data); e != nil {
		return e
	}
	if _, e := c.conn.Write([]byte(util.EOT)); e != nil {
		return e
	}
	return readOk(c.conn)
}

func (c *client) sendFile(numFiles int, files []file) error {
	var i int = 0
outer:
	for {
		for _, f := range files {
			if i == numFiles {
				break outer
			}
			if e := write(c.conn, map[string]interface{}{REQUEST: SEND, project.TYPE: f.tipe, db.NAME: f.name, db.PKG: f.pkg, db.TIME: util.CurMilis()}); e != nil {
				return e
			}
			if e := readOk(c.conn); e != nil {
				return e
			}
			if _, e := c.conn.Write(f.data); e != nil {
				return e
			}
			if _, e := c.conn.Write([]byte(util.EOT)); e != nil {
				return e
			}
			if e := readOk(c.conn); e != nil {
				return e
			}
			i++
		}
	}
	return nil
}

func write(c net.Conn, data interface{}) error {
	if e := util.WriteJSON(c, data); e != nil {
		return e
	}
	_, e := c.Write([]byte(util.EOT))
	return e
}

func readOk(c net.Conn) error {
	d, e := util.ReadData(c)
	if e == nil && !bytes.HasPrefix(d, []byte(OK)) {
		return fmt.Errorf("unexpected reply %s", string(d))
	}
	return e
}

func loadZip(fileNum int) ([]byte, error) {
	data := make(map[string][]byte)
	start := 1377870393875
	for i := 0; i < int(fileNum); i++ {
		name := fmt.Sprintf("triangle_Triangle.java_%d_c", start+i)
		data[name] = fileData
	}
	return util.ZipMap(data)
}

func (c *client) run(port, numFiles int, files []file) error {
	a, e := c.login(port)
	if e != nil {
		return e
	}
	if e = c.create(a); e != nil {
		return e
	}
	if e = c.send(numFiles, files); e != nil {
		return e
	}
	return c.logout()
}

func testReceive(spawner *clientSpawner) error {
	errChan := make(chan error)
	nU := len(spawner.users)
	ok := true
	cli, ok := spawner.spawn()
	for ok {
		go func(c *client) {
			errChan <- c.run(spawner.rport, spawner.numFiles, spawner.files)
		}(cli)
		cli, ok = spawner.spawn()
	}
	done := 0
	var e error
	for done < nU {
		if e = <-errChan; e != nil {
			return e
		}
		done++
	}
	return mq.WaitIdle()
}

func testFiles(t *testing.T, nF, nU, port int, mode string, files []file) {
	fmt.Printf("testing files %d files %d users %s mode\n", nF, nU, mode)
	ext := "_" + strconv.Itoa(int(port))
	db.Setup(db.TEST_CONN + ext)
	db.DeleteDB(db.TEST_DB + ext)
	db.Setup(db.TEST_CONN + ext)
	defer db.DeleteDB(db.TEST_DB + ext)
	users, e := addData(nU)
	if e != nil {
		t.Error(e)
	}
	go monitor.Start()
	go processor.Serve(processor.MAX_PROCS)
	receive(port)
	spawner := &clientSpawner{
		mode:     mode,
		files:    files,
		users:    users,
		numFiles: nF,
		rport:    port,
	}
	if e = testReceive(spawner); e != nil {
		t.Error(e)
	}
	if e = processor.Shutdown(); e != nil {
		t.Error(e)
	}
	fmt.Printf("tested files %d files %d users %s mode\n", nF, nU, mode)
}

func TestFile(t *testing.T) {
	files := []file{{"Triangle.java", "triangle", project.SRC, fileData}}
	testFiles(t, 1, 1, 8005, project.FILE_MODE, files)
	testFiles(t, 2, 3, 8020, project.FILE_MODE, files)
	files = append(files, file{"UserTests.java", "testing", project.TEST, userTestData})
	testFiles(t, 2, 1, 8030, project.FILE_MODE, files)
	testFiles(t, 4, 3, 8040, project.FILE_MODE, files)
	zipData, err := loadZip(1)
	if err != nil {
		t.Error(err)
	}
	zips := []file{{"Triangle.java", "triangle", project.ARCHIVE, zipData}}
	testFiles(t, 1, 1, 8050, project.ARCHIVE_MODE, zips)
	zipData, err = loadZip(5)
	if err != nil {
		t.Error(err)
	}
	testFiles(t, 3, 2, 8010, project.ARCHIVE_MODE, zips)
}

func benchmarkFiles(b *testing.B, nF, nU, nS, nM, port int, mode string, files []file) {
	servers := make([]*processor.Server, nS)
	var err error
	for i := 0; i < int(nS); i++ {
		servers[i], err = processor.NewServer(processor.MAX_PROCS)
		if err != nil {
			b.Error(err)
		}
		go servers[i].Serve()
	}
	monitors := make([]*monitor.M, nS)
	for i := 0; i < int(nM); i++ {
		monitors[i], err = monitor.New()
		if err != nil {
			b.Error(err)
		}
		go monitors[i].Start()
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

/*
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
*/

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
