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


func loginClient(port, uname, pword string) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return clientLogin(conn, uname, pword)
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


func sendTest(conn net.Conn) error {
	fmap := map[string]interface{}{REQ:SEND, submission.PROJECT:"project", submission.LANG: "lang", submission.NAMES: []string{"test0","test1","test2"}}
	err := writeJson(conn, fmap)
	if err != nil {
		return err
	}
	err = checkMessage(conn, OK)
	if err != nil {
		return err
	}
	err = writeData(conn, testData)
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
	return checkMessage(conn, OK)
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

func loginSubServer(port string) error{
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
	err = loginSubServer(port)
	if err != nil{
			t.Error(err)
	}
	go func() {
		err = loginClient(port, uname, "")
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginSubServer(port)
	if err == nil{
		t.Error("Expected error")
	}
	go func() {
		err = loginClient(port, "", pword)
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginSubServer(port)
	if err == nil{
		t.Error("Expected error")
	}
}

func readSubClient(port string, fileChan chan *submission.File) error{
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


func readSubServer(port string, fileChan chan *submission.File) error{
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
		err := readSubClient(port, fileChan)
		if err != nil{
			t.Error(err)
		}
	}()
	err := readSubServer(port, fileChan)
	if err != nil {
		t.Error(err)
	}
}

func handleSubClient(port, uname,pword string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
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

func handleSubServer(port string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
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
		err = handleSubClient(port,uname,pword, subChan, fileChan)
		if err != nil {
			t.Error(err)
		}
	}()
	err = handleSubServer(port, subChan, fileChan)
	if err != nil {
		t.Error(err)
	}
}



func startSubClient(port, uname,pword string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
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
	return checkMessage(conn, OK) 
}

func startSubServer(port string, subChan chan *submission.Submission, fileChan chan *submission.File) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &SubmissionHandler{SubChan: subChan, FileChan: fileChan}
	handler.Start(sconn)
	return nil
}


func TestSubStart(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname, pword := "unames", "pwords"
	hash, salt := util.Hash(pword)
	u := user.NewUser(uname, hash, salt)
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "4000"
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go func() {
		err = startSubClient(port,uname,pword, subChan, fileChan)
		if err != nil {
			t.Error(err)
		}
	}()
	err = startSubServer(port, subChan, fileChan)
	if err != nil {
		t.Error(err)
	}
}


func loginTestClient(port, uname, pword string) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return clientLogin(conn, uname, pword)
}

func loginTestServer(port string) error{
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &TestHandler{Conn: sconn}
	return handler.Login()
}

func TestTestLogin(t *testing.T){
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname,pword := "unametl","pwordtl"
	hash, salt := util.Hash(pword)
	u := &user.User{uname, hash, salt, user.T_SUB}
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "9001"
	go func() {
		err = loginClient(port, uname, pword)
		if err != nil{
			t.Error(err)
		}
	}()
	err = loginTestServer(port)
	if err != nil{
			t.Error(err)
	}
	go func() {
		err = loginClient(port, uname, "")
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginTestServer(port)
	if err == nil{
		t.Error("Expected error")
	}
	go func() {
		err = loginClient(port, "", pword)
		if err == nil{
			t.Error("Expected error")
		}
	}()
	err = loginTestServer(port)
	if err == nil{
		t.Error("Expected error")
	}
}

func readTestClient(port string) error{
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	return sendTest(conn)
}


func readTestServer(port string) error{
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &TestHandler{Conn: sconn}
	err = handler.Read()
	if err != nil {
		return err
	}
	expected := submission.NewTest("project", "lang",[]string{"test0","test1","test2"}, testData, fileData)
	expected.Id = handler.Test.Id
	if !expected.Equals(handler.Test){
		return fmt.Errorf("Tests not equivalent")
	}
	return nil
}

func TestTestRead(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	port := "7001"
	go func() {
		err := readTestClient(port)
		if err != nil{
			t.Error(err)
		}
	}()
	err := readTestServer(port)
	if err != nil {
		t.Error(err)
	}
}

func handleTestClient(port, uname,pword string) error {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	err = clientLogin(conn, uname, pword)
	if err != nil {
		return err
	}
	return sendTest(conn)
}

func handleTestServer(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &TestHandler{Conn: sconn}
	return handler.Handle()
}

func TestTestHandle(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname, pword := "unameth", "pwordth"
	hash, salt := util.Hash(pword)
	u := &user.User{uname, hash, salt, user.T_SUB}
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "6001"
	go func() {
		err = handleTestClient(port,uname,pword)
		if err != nil {
			t.Error(err)
		}
	}()
	err = handleTestServer(port)
	if err != nil {
		t.Error(err)
	}
}



func startTestClient(port, uname,pword string) error {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer conn.Close()
	err = clientLogin(conn, uname, pword)
	if err != nil {
		return err
	}
	err = sendTest(conn)
	if err != nil {
		return err
	}
	return checkMessage(conn, OK) 
}

func startTestServer(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	sconn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer sconn.Close()
	handler := &TestHandler{}
	handler.Start(sconn)
	return nil
}


func TestTestStart(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	uname, pword := "unamets", "pwordts"
	hash, salt := util.Hash(pword)
	u := &user.User{uname, hash, salt, user.T_SUB}
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	port := "4001"
	go func() {
		err = startTestClient(port,uname,pword)
		if err != nil {
			t.Error(err)
		}
	}()
	err = startTestServer(port)
	if err != nil {
		t.Error(err)
	}
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

var testData = []byte(`Despicã-mi oglinda de dincolo,
Cu a ochilor ascutime...
Chip ce neclintit se asterne-n sine.
Tintesc lumile, întorc luminile
Liniste lãuntricã se dezveleste
Dincolo de a mintii serpuire.

Despic cãrarea nerostirii, prãbusitã-n nemiscare
Umblet prin suflet pe sfori de nepãsit.
Cuvinte surde croiesc prin minte,
Nori de vorbe îndeasã neîncetat
Zumzãind cu sunet aspru.
Cãtre stînga rãsucesc privirea
Ce strãpunge scutul lãuntric, tãcere
Cuprins de eterna negîndire.

Calmã adiere ce de dincolo mã trage
De-a curmezis de lume, purtatu-s de Zefir
De la începuturi cãtre Nadir
Întelept domnit, alunecînd ca prin vis,
Albind întunecatii mei strãmosi de stîncã.

Despic puterea,
Cu a ochilor asprime,
În mãduva bradului de veci,
Si cum stau înaintea-mi,
Înghitit neantului
Înãuntru-mi privesc
Asa aproape de mine,
Strãluminat de cel ochi, mãiastru
Si insumi sunt.`)