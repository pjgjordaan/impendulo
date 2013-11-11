package receiver

import (
	"bytes"
	"encoding/json"
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

var (
	nU, nF, rport, pport = 1, 2, 8010, 8020
)

func init() {
	fmt.Print()
	util.SetErrorLogging("a")
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

func login(uname, pword, mode string, port int) (conn net.Conn, infos []*ProjectInfo, err error) {
	conn, err = net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	req := map[string]interface{}{
		REQ:          LOGIN,
		db.USER:      uname,
		db.PWORD:     pword,
		project.MODE: mode,
	}
	err = write(conn, req)
	if err != nil {
		return
	}
	read, err := util.ReadData(conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(read, &infos)
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

func create(conn net.Conn, projectId bson.ObjectId) (sub *project.Submission, err error) {
	req := map[string]interface{}{
		REQ:          NEW,
		db.PROJECTID: projectId,
		db.TIME:      util.CurMilis(),
	}
	err = write(conn, req)
	if err != nil {
		return
	}
	read, err := util.ReadData(conn)
	if err != nil {
		return
	}
	err = json.Unmarshal(read, &sub)
	return
}

func logout(conn net.Conn) (err error) {
	req := map[string]interface{}{
		REQ: LOGOUT,
	}
	err = write(conn, req)
	if err != nil {
		return
	}
	err = readOk(conn)
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

type sender func(net.Conn, []byte, int) error

func sendArchive(conn net.Conn, zipped []byte, nF int) (err error) {
	req := map[string]interface{}{
		REQ: SEND,
	}
	err = write(conn, req)
	if err != nil {
		return
	}
	err = readOk(conn)
	if err != nil {
		return
	}
	_, err = conn.Write(zipped)
	if err != nil {
		return
	}
	_, err = conn.Write([]byte(util.EOT))
	if err != nil {
		return
	}
	err = readOk(conn)
	return
}

func sendFile(conn net.Conn, file []byte, nF int) (err error) {
	req := map[string]interface{}{
		REQ:          SEND,
		project.TYPE: project.SRC,
		db.NAME:      "Triangle.java",
		db.PKG:       "triangle",
	}
	for i := 0; i < nF; i++ {
		req[db.TIME] = util.CurMilis()
		err = write(conn, req)
		if err != nil {
			return
		}
		err = readOk(conn)
		if err != nil {
			return
		}
		_, err = conn.Write(file)
		if err != nil {
			return
		}
		_, err = conn.Write([]byte(util.EOT))
		if err != nil {
			return
		}
		err = readOk(conn)
		if err != nil {
			return
		}
	}
	return
}

func TestReceive(t *testing.T) {
	db.Setup(db.TEST_CONN)
	db.DeleteDB(db.TEST_DB)
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	users, err := addData(nU)
	if err != nil {
		t.Error(err)
	}
	processing.SetClientAddress("", pport)
	receive(rport)
	testReceive(t, project.FILE_MODE, fileData, sendFile, users)
	zipData, err := loadZip(nF)
	if err != nil {
		t.Error(err)
	}
	testReceive(t, project.ARCHIVE_MODE, zipData, sendArchive, users)
}

func testReceive(t *testing.T, mode string, data []byte, send sender, users map[string]string) {
	go func() {
		doneChan := make(chan struct{})
		for uname, pword := range users {
			go func(u string) {
				conn, infos, err := login(u, pword, mode, rport)
				if err != nil {
					t.Error(err)
				}
				_, err = create(conn, infos[0].Project.Id)
				if err != nil {
					t.Error(err)
				}
				err = send(conn, data, nF)
				if err != nil {
					t.Error(err)
				}
				err = logout(conn)
				if err != nil {
					t.Error(err)
				}
				doneChan <- struct{}{}
			}(uname)
		}
		done := 1
		for done < len(users) {
			<-doneChan
			done++
		}
		time.Sleep(1 * time.Second)
		processing.Shutdown()
	}()
	processing.Serve(pport, 10)
}

/*
func BenchmarkArchive(b *testing.B) {
	nU, nF := 1, 2
	db.Setup(db.TEST_CONN)
	db.DeleteDB(db.TEST_DB)
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	util.SetErrorLogging("a")
	processing.SetClientAddress("", 8040)
	users, err := addData(nU)
	if err != nil {
		b.Error(err)
	}
	zipped, err := loadZip(nF)
	if err != nil {
		b.Error(err)
	}
	receive()
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		go func() {
			doneChan := make(chan struct{})
			for uname, pword := range users {
				go func(u string) {
					conn, infos, err := login(u, pword)
					if err != nil {
						b.Error(err)
					}
					_, err = create(conn, infos[0].Project.Id)
					if err != nil {
						b.Error(err)
					}
					err = sendArchive(conn, zipped)
					if err != nil {
						b.Error(err)
					}
					err = logout(conn)
					if err != nil {
						b.Error(err)
					}
					doneChan <- struct{}{}
				}(uname)
			}
			done := 0
			for done < len(users) {
				<-doneChan
				done++
			}
			processing.Shutdown()
		}()
		processing.Serve(8040, 10)
	}
}*/

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
