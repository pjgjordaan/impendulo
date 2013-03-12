package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"io"
	"net"
	"errors"
	"strconv"
)
func getPath(fname string, c *client.Client)(file string, err error){
	if c.Mode == client.ONSAVE{
		file = c.Project + utils.SEP + c.Name + utils.SEP + fname
	} else if c.Mode == client.ONSTOP{
		file = c.Project + utils.SEP + c.Name +"_" +c.Project+strconv.Itoa(c.ProjectNum)+".zip"
	} else{
		err = errors.New("Unknown send mode: "+c.Mode)
	}
	return file, err
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
		clientError(conn, "Unexpected error", token, err)
	} else {
		c := tokens[token]
		file, err := getPath(fname, c)
		if err != nil{
			clientError(conn, "Path error: "+err.Error(), token, err)
		}else{		
			err = utils.WriteFile(file, buffer)
			if err != nil {
				clientError(conn, "Write error: "+err.Error(), token, err)
			} else {
				utils.Log("Successfully wrote for: ", c.Name)
				conn.Close()
			}
		}
	}
}

func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		clientError(conn, "Client connection error", "", err)
	} else {
		handleRequest(buffer[:bytesRead], conn)
	}
}

func handleRequest(data []byte, conn net.Conn) {
	var holder interface{}
	err := json.Unmarshal(data, &holder)
	if err != nil {
		clientError(conn, "Invalid connection request: "+string(data), "", err)
	} else {
		jobj := holder.(map[string]interface{})
		val, err := utils.JSONValue(jobj, "TYPE")
		if err != nil {
			clientError(conn, "Invalid request type "+string(data), "", err)
		} else if val == "LOGIN" {
			handleLogin(jobj, conn)
		} else if val == "ZIP" {
			handleZip(jobj, conn)
		} else if val == "SEND" {
			handleSend(jobj, conn)
		} else if val == "LOGOUT" {
			handleLogout(jobj, conn)
		}
	}
}

func handleLogin(jobj map[string]interface{}, conn net.Conn) {
	uname, erru := utils.JSONValue(jobj, "USERNAME")
	pword, errw := utils.JSONValue(jobj, "PASSWORD")
	project, errp := utils.JSONValue(jobj, "PROJECT")
	mode, errm := utils.JSONValue(jobj, "MODE")
	if erru != nil || errw != nil || errp != nil || errm != nil{
		clientError(conn, "Error retrieving JSON value from request", "", nil)
	} else {
		num, err := getProjectNum(uname, pword, project) 
		if err == nil {
			c := client.NewClient(uname, project, num, mode)
			token := getToken(c)
			err := utils.MkDir(c.Project+utils.SEP+c.Name)
			if err != nil {
				clientError(conn, "Error creating project: "+err.Error(), token, err)
			} else {
				conn.Write([]byte("TOKEN:" + token))
			}
		} else {
			clientError(conn, "Invalid username, password or project "+uname+" "+pword + " " +project, "", err)
		}
	}
}

func handleSend(jobj map[string]interface{}, conn net.Conn) {
	token, errt := utils.JSONValue(jobj, "TOKEN")
	fname, errf := utils.JSONValue(jobj, "FILENAME")
	if errt != nil || errf != nil {
		clientError(conn, "Error retrieving JSON value from request", "", nil)
	} else {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			FileReader(conn, token, fname)
		} else {
			clientError(conn, "Invalid token "+token, "", nil)
		}
	}
}

func handleZip(jobj map[string]interface{}, conn net.Conn) {
	token, errt := utils.JSONValue(jobj, "TOKEN")
	if errt != nil {
		clientError(conn, "Error retrieving JSON value from request", "", errt)
	} else {
		if tokens[token] != nil {
			c := tokens[token]
			err := utils.ZipProject(c)
			if err == nil{
				err = utils.Remove(c.Project+utils.SEP+c.Name)
			}
			conn.Write([]byte("ACCEPT"))
			if err != nil {
				clientError(conn, "Error creating zip file "+token, token, err)
			} else {
				conn.Close()
			}
		} else {
			clientError(conn, "Invalid token "+token, "", nil)
		}
	}
}

func handleLogout(jobj map[string]interface{}, conn net.Conn) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err != nil {
		clientError(conn, "Error retrieving token from request", "", err)
	} else {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			conn.Close()
			delete(tokens, token)
		} else {
			clientError(conn, "Invalid token "+token, "", nil)
		}
	}
}

func clientError(conn net.Conn, msg string, token string, err error) {
	utils.Log(msg, err)
	conn.Write([]byte(msg))
	conn.Close()
	if token != "" {
		c := tokens[token]
		utils.Remove(c.Project+utils.SEP+c.Name)
		delete(tokens, token)
	}
}

func getProjectNum(uname, pword, project string) (int, error) {
	info, err := db.Read(uname)
	num := -1
	if err == nil {
		if info.Password == pword {
			info.Projects[project] ++
			num = info.Projects[project]
			err = db.Add(uname, info)		
		} else {
			err = errors.New("Invalid password: "+pword)
		}
	}
	return num, err
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

func Run(address string, port string) {
	utils.Log("Server Started")
	service := address + ":" + port
	tcpAddr, error := net.ResolveTCPAddr("tcp", service)
	if error != nil {
		utils.Log("Error: Could not resolve address")
	} else {
		netListen, error := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if error != nil {
			utils.Log(error)
		} else {
			defer netListen.Close()
			for {
				conn, error := netListen.Accept()
				if error != nil {
					utils.Log("Client error: ", error)
				} else {
					go ConnHandler(conn)
				}
			}
		}
	}
}
