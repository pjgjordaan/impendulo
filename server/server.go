package server

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intlola/client"
	"io"
	"io/ioutil"
	"net"
	"os"
"github.com/peterbourgon/diskv"
)

const sep = string(os.PathSeparator)

func WriteFile(file string, data *bytes.Buffer) error {
	Log("Writing to: ", file)
	err := ioutil.WriteFile(file, data.Bytes(), 0666)
	return err
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}

func createUserProject(client *client.Client) error {
	os.Mkdir(client.Project, 0777)
	return os.Mkdir(client.Project+sep+client.Name, 0777)
}

func ZipDir(dir, fname string) (err error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	finfos, err := ioutil.ReadDir(dir)
	if err == nil {
		for _, file := range finfos {
			if !file.IsDir() {
				f, err := w.Create(file.Name())
				if err != nil {
					break
				}
				contents, err := ioutil.ReadFile(dir+sep+file.Name())
				if err != nil {
					break
				}
				_, err = f.Write(contents)
				if err != nil {
					break
				}
			}
		}
	
	}
	w.Close()
	if err == nil {
		err = WriteFile(dir+sep+fname, buf)
	} 
	return err
}

func FileReader(conn net.Conn, token string, fname string) {
	buffer := new(bytes.Buffer)
	p := make([]byte, 2048)
	bytesRead, err := conn.Read(p)
	for err == nil {
		buffer.Write(p[:bytesRead])
		bytesRead, err = conn.Read(p)
	}
	if err != io.EOF {
		clientError(conn, "Unexpected error: "+err.Error(), token)
	} else {
		c := tokens[token]
		file := c.Project + sep + c.Name + sep + fname
		err = WriteFile(file, buffer)
		if err != nil {
			clientError(conn, "Write error: "+err.Error(), token)
		} else {
			Log("Successfully wrote for: ", c.Name)
			conn.Close()
		}
	}
}

func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		clientError(conn, "Client connection error: "+error.Error(), "")
	} else {
		handleRequest(buffer[:bytesRead], conn)
	}
}

func handleRequest(data []byte, conn net.Conn) {
	var holder interface{}
	err := json.Unmarshal(data, &holder)
	if err != nil {
		clientError(conn, "Invalid connection request: "+string(data), "")
	} else {
		jobj := holder.(map[string]interface{})
		val, err := getJSONValue(jobj, "TYPE")
		if err != nil {
			clientError(conn, "Invalid request type "+string(data)+" error "+err.Error(), "")
		} else if val == "LOGIN" {
			handleLogin(jobj, conn)
		} else if val == "ZIP" {
			handleZip(jobj, conn)
		} else if val == "SEND" {
			handleSend(jobj, conn)
		} else if val == "LOGOUT"{
			handleLogout(jobj, conn)
		}
	}
}

func handleLogin(jobj map[string]interface{}, conn net.Conn) {
	uname, erru := getJSONValue(jobj, "USERNAME")
	pword, errw := getJSONValue(jobj, "PASSWORD")
	project, errp := getJSONValue(jobj, "PROJECT")
	if erru != nil || errw != nil || errp != nil {
		clientError(conn, "Error retrieving JSON value from request", "")
	} else {
		if validate(uname, pword) {
			client := client.NewClient(uname, project)
			token := getToken(client)
			err := createUserProject(client)
			if err != nil {
				clientError(conn, "Error creating project: "+err.Error(), token)
			} else {
				
				conn.Write([]byte("TOKEN:"+token))
			}
		} else {
			clientError(conn, "Invalid username or password "+uname+" "+pword, "")
		}
	}
}

func handleSend(jobj map[string]interface{}, conn net.Conn) {
	token, errt := getJSONValue(jobj, "TOKEN")
	fname, errf := getJSONValue(jobj, "FILENAME")
	if errt != nil || errf != nil {
		clientError(conn, "Error retrieving JSON value from request", "")
	} else {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			FileReader(conn, token, fname)
		} else {
			clientError(conn, "Invalid token "+token, "")
		}
	}
}

func handleZip(jobj map[string]interface{}, conn net.Conn) {
	token, errt := getJSONValue(jobj, "TOKEN")
	fname, errf := getJSONValue(jobj, "FILENAME")
	if errt != nil || errf != nil {
		clientError(conn, "Error retrieving JSON value from request", "")
	} else {
		if tokens[token] != nil {
			c := tokens[token]
			dir := c.Project + sep + c.Name
			err := ZipDir(dir, fname)
			conn.Write([]byte("ACCEPT"))
			if err != nil {
				clientError(conn, "Error creating zip file "+token, token)
			} else{ 
				conn.Close()
			}
		} else {
			clientError(conn, "Invalid token "+token, "")
		}
	}
}

func handleLogout(jobj map[string]interface{}, conn net.Conn) {
	token, err := getJSONValue(jobj, "TOKEN")
	if err != nil {
		clientError(conn, "Error retrieving token from request", "")
	} else {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			conn.Close()
			delete(tokens, token)
		} else {
			clientError(conn, "Invalid token "+token, "")
		}
	}
}

func clientError(conn net.Conn, msg string, token string) {
	Log(msg)
	conn.Write([]byte(msg))
	conn.Close()
	if token != "" {
		delete(tokens, token)
	}
}

var users map[string]string

func validate(uname, pword string) bool {
	valbytes, err := db.Read(uname)
	if len(valbytes) == 0 || err != nil{
		db.Write(uname, [] byte(pword))
		valbytes, _ = db.Read(uname)
	} 
	val := string(valbytes)
	return val == pword
}

func getJSONValue(jobj map[string]interface{}, key string) (string, error) {
	ival, err := jobj[key]
	if !err {
		return "", errors.New(key + " not found in JSON Object")
	}
	switch val := ival.(type) {
	case string:
		return val, nil
	}
	return "", errors.New("Invalid type in JSON parameter")
}

var tokens map[string]*client.Client

const STRLEN = 80

func genToken() string {
	b := make([]byte, STRLEN)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}
func getToken(c *client.Client) string {
	if tokens == nil {
		tokens = make(map[string]*client.Client)
	}
	tok := genToken()
	for tokens[tok] != nil {
		tok = genToken()
	}
	tokens[tok] = c
	return tok
}

var db *diskv.Diskv
func initDB(){
	flatTransform := func(s string) []string{return []string{""}}
	db = diskv.New(diskv.Options{
		BasePath: "db",
		Transform: flatTransform,
	CacheSizeMax: 1024 * 1024,})
}
func Run(address string, port string) {
	initDB()
	Log("Server Started")
	service := address + ":" + port
	tcpAddr, error := net.ResolveTCPAddr("tcp", service)
	if error != nil {
		Log("Error: Could not resolve address")
	} else {
		netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if error != nil {
			Log(error)
		} else {
			defer netListen.Close()
			for {
				conn, error := netListen.Accept()
				if error != nil {
					Log("Client error: ", error)
				} else {
					go ConnHandler(conn)
				}
			}
		}
	}
}
