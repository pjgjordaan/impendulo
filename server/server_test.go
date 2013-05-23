package server

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
//	"labix.org/v2/mgo/bson"
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

func clientLogin(conn net.Conn, uname, pword string) error {
	umap := map[string]interface{}{REQ: LOGIN, submission.USER: uname, user.PWORD: pword, submission.PROJECT: "project", submission.MODE: submission.FILE_MODE, submission.LANG: "java"}
	err := writeJson(conn, umap)
	if err != nil {
		return err
	}
	return checkMessage(conn, OK)
}

func sendFile(conn net.Conn, fileChan chan *submission.File) error {
	err := writeJson(conn, map[string]interface{}{REQ: SEND})
	if err != nil{
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	err = writeData(conn, fileData)
	if err != nil {
		return err
	}
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

func writeData(conn net.Conn, data []byte)error{
	_, err := conn.Write(data)
	if err != nil {
		return err
	}
	_, err = conn.Write(eof)
	if err != nil {
		return err
	}
	return nil
}

func writeJson(conn net.Conn, data map[string]interface{}) error{
	marshalled, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = writeData(conn, marshalled)
	if err != nil {
		return err
	}
	return err
}

func basicClient(port string, doneChan chan bool)error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	err = writeJson(conn, map[string]interface{}{"A": "2", "B": " H"})
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	err = writeData(conn, fileData)
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	doneChan <- true
	return nil		
}

func TestRun(t *testing.T){
	doneChan := make(chan bool)
	n := 100
	port := "8100"
	go Run(port, new(BasicSpawner))
	go func() {
		for i := 0; i < n; i ++{
			go func() {
				err := basicClient(port, doneChan)
				if err != nil{
					t.Error(err)
				}
			}()
		}
	}()
	count := 0
	for _ = range doneChan{
		count ++
		if count == 100{
			break
		}
	}
}

func TestReadSubmission(t *testing.T) {
	expected := submission.NewSubmission("project", "uname",submission.FILE_MODE, "java")
	umap := map[string]interface{}{submission.USER: "uname", user.PWORD: "pword", submission.PROJECT: "project", submission.MODE: submission.FILE_MODE, submission.LANG: "java"}
	handler := new(SubmissionHandler)
	err := handler.ReadSubmission(umap)
	if err != nil {
		t.Error(err)
	}
	expected.Id = handler.Submission.Id 
	expected.Time = handler.Submission.Time
	if !expected.Equals(handler.Submission){
		t.Error("Submissions don't match", expected, handler.Submission)
	}
}

func loginClient(port, uname, pword string) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return clientLogin(conn, uname, pword)
}

func loginServer(port string) error{
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &SubmissionHandler{Conn: sconn}
	return handler.Login()
}

func TestSubLogin(t *testing.T){
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname,pword := "unamel","pwordl"
	hash, salt := util.Hash(pword)
	u := user.NewUser(uname, hash, salt)
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "9000"
	go func() {
		err = loginClient(port, uname, pword)
		if err != nil{
			t.Error(err)
		}
	}()
	err = loginServer(port)
	if err != nil{
			t.Error(err)
	}
	go func() {
		err = loginClient(port, uname, "")
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginServer(port)
	if err == nil{
		t.Error("Expected error")
	}
	go func() {
		err = loginClient(port, "", pword)
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginServer(port)
	if err == nil{
		t.Error("Expected error")
	}
}

func readClient(port string, fileChan chan *submission.File) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmap := map[string]interface{}{REQ:SEND}
	err = writeJson(conn, fmap)
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	err = writeData(conn, fileData)
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	recv := <-fileChan
	if !reflect.DeepEqual(fileData, recv.Data) {
		return fmt.Errorf("Data not the same")
	}
	return nil
}


func readServer(port string, fileChan chan *submission.File) error{
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	sub := submission.NewSubmission("","","","")
	handler := &SubmissionHandler{Conn: sconn, Submission: sub, FileChan: fileChan}
	return handler.Read()
}

func TestSubRead(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	fileChan := make(chan *submission.File)
	port := "7000"
	go func() {
		err := readClient(port, fileChan)
		if err != nil{
			t.Error(err)
		}
	}()
	err := readServer(port, fileChan)
	if err != nil {
		t.Error(err)
	}
}

func endClient(port string) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return checkMessage(conn, OK)
}

func endServer(port string) error{
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	handler := &BasicHandler{Conn: sconn}
	handler.End(nil)
	return nil
}

func TestEnd(t *testing.T) {
	port := "2000"
	go func() {
		err := endClient(port)
		if err != nil{
			t.Error(err)
		}
	}()
	err := endServer(port)
	if err != nil{
		t.Error(err)
	}
}

func handleClient(port, uname,pword string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	err = clientLogin(conn, uname, pword)
	if err != nil {
		return err
	}
	<- subChan
	for i := 0; i < 10; i++ {
		err = sendFile(conn, fileChan)
		if err != nil {
			return err
		}
	}
	logout := map[string]interface{}{REQ: LOGOUT}
	err = writeJson(conn, logout)
	if err != nil {
		return err
	}
	<- subChan
	return nil
}

func handleServer(port string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &SubmissionHandler{Conn: sconn, SubChan: subChan, FileChan: fileChan}
	return handler.Handle()
}

func TestSubHandle(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname, pword := "unameh", "pwordh"
	hash, salt := util.Hash(pword)
	u := user.NewUser(uname, hash, salt)
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "6000"
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go func() {
		err = handleClient(port,uname,pword, subChan, fileChan)
		if err != nil {
			t.Error(err)
		}
	}()
	err = handleServer(port, subChan, fileChan)
	if err != nil {
		t.Error(err)
	}
}

/*
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
*/

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
