package server

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"labix.org/v2/mgo/bson"
	"net"
	"reflect"
	"testing"
)

func checkMessage(conn net.Conn, check string) error {
	data := make([]byte, len([]byte(check)))
	n, err := conn.Read(data)
	if err != nil {
		return err
	}
	if n != len(check) {
		return fmt.Errorf("Did not read all data from connection")
	}
	msg := string(data)
	if msg != OK {
		return fmt.Errorf("Got unexpected message %s", msg)
	}
	return nil
}

func clientLogin(conn net.Conn, subChan chan *submission.Submission) error {
	umap := map[string]interface{}{REQ: LOGIN, submission.USER: "uname", user.PWORD: "pword", submission.PROJECT: "project", submission.MODE: submission.FILE_MODE, submission.LANG: "java"}
	ubytes, err := json.Marshal(umap)
	if err != nil {
		return err
	}
	conn.Write(ubytes)
	conn.Write(eof)
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	<-subChan
	return nil
}

func clientLogout(conn net.Conn) error {
	logout := map[string]interface{}{REQ: LOGOUT}
	lbytes, err := json.Marshal(logout)
	if err != nil {
		return err
	}
	conn.Write(lbytes)
	conn.Write(eof)
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	return nil
}

func sendFile(conn net.Conn, fileChan chan *submission.File) error {
	fmap := map[string]interface{}{REQ: SEND}
	fbytes, err := json.Marshal(fmap)
	if err != nil {
		return err
	}
	conn.Write(fbytes)
	conn.Write(eof)
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	conn.Write(fileData)
	conn.Write(eof)
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	recv := <-fileChan
	if !reflect.DeepEqual(fileData, recv.Data) {
		return fmt.Errorf("Data not equal")
	}
	return nil
}

func TestCreateSubmission(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	hash, salt := util.Hash("pword")
	u := user.NewUser("uname", hash, salt)
	umap := map[string]interface{}{submission.USER: "uname", user.PWORD: "pword", submission.PROJECT: "project", submission.MODE: submission.FILE_MODE, submission.LANG: "java"}
	_, err := CreateSubmission(umap)
	if err == nil {
		t.Error("should get not found error")
	}
	err = db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	_, err = CreateSubmission(umap)
	if err != nil {
		t.Error(err)
	}
	_, err = CreateSubmission(map[string]interface{}{})
	if err == nil {
		t.Error("error expected")
	}
	umap[user.PWORD] = "a"
	_, err = CreateSubmission(umap)
	if err == nil {
		t.Error("error expected")
	}
}

func TestEndSession(t *testing.T) {
	go func() {
		conn, err := net.Dial("tcp", "localhost:8000")
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		err = checkMessage(conn, OK)
		if err != nil {
			t.Error(err)
		}
	}()
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		t.Error(err)
	}
	sconn, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	EndSession(sconn, nil)

}

func TestLogin(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	hash, salt := util.Hash("pword")
	u := user.NewUser("uname", hash, salt)
	umap := map[string]interface{}{submission.USER: "uname", user.PWORD: "pword", submission.PROJECT: "project", submission.MODE: submission.FILE_MODE, submission.LANG: "java"}
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	go func() {
		conn, err := net.Dial("tcp", "localhost:9000")
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		err = checkMessage(conn, OK)
		if err != nil {
			t.Error(err)
		}
	}()
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		t.Error(err)
	}
	sconn, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	defer sconn.Close()
	_, err = Login(umap, sconn)
	if err != nil {
		t.Error(err)
	}

}

func TestProcessFile(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	fileChan := make(chan *submission.File)
	subId := bson.NewObjectId()
	fmap := map[string]interface{}{}
	go func() {
		conn, err := net.Dial("tcp", "localhost:7000")
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		err = checkMessage(conn, OK)
		if err != nil {
			t.Error(err)
		}
		conn.Write(fileData)
		conn.Write(eof)
		err = checkMessage(conn, OK)
		if err != nil {
			t.Error(err)
		}
		recv := <-fileChan
		if !reflect.DeepEqual(fileData, recv.Data) {
			t.Error(fileData, "!=", recv.Data)
		}
	}()
	ln, err := net.Listen("tcp", ":7000")
	if err != nil {
		t.Error(err)
	}
	sconn, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	defer sconn.Close()
	err = ProcessFile(subId, fmap, sconn, fileChan)
	if err != nil {
		t.Error(err)
	}

}

func TestConnHandler(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	hash, salt := util.Hash("pword")
	u := user.NewUser("uname", hash, salt)
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	go func() {
		conn, err := net.Dial("tcp", "localhost:6000")
		if err != nil {
			t.Error(err)
		}
		defer conn.Close()
		err = clientLogin(conn, subChan)
		if err != nil {
			t.Error(err)
		}
		for i := 0; i < 10; i++ {
			err = sendFile(conn, fileChan)
			if err != nil {
				t.Error(err)
			}
		}
		err = clientLogout(conn)
		if err != nil {
			t.Error(err)
		}
	}()
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		t.Error(err)
	}
	sconn, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	defer sconn.Close()
	ConnHandler(sconn, subChan, fileChan)
}

var eof = []byte("eof")

var fileData = []byte(`Don stepped outside 
It feels good to be alone
He wished he was drunk
He thought about something he said
And how stupid it had sounded
He should forget about it
He decided to piss, but he couldn't
(A plane passes silently overhead)

The streetlights, and the buds on the trees, were still

It finally came, he took a deep breath
It made him feel strong, and determined
To go back inside

The light
Their backs
The conversations
The couples, romancing, so natural
His friends stare
With eyes like the heads of nails
The others
Glances
With amusement
With evasion
With contempt
So distant
With malice
For being a sty in their engagement
Like swimming underwater in the darkness
Like walking through an empty house
Speaking to an imaginary audience

And being watched from outside by
Someone without a key

He could not dance to anything
Don left
And drove
And howled
And laughed
At himself
He felt he knew what that was

Don woke up
And looked at the night before
He knew what he had to do
He was responsible
In the mirror
He saw his friend`)
