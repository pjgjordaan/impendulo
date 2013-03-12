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
		clientError(conn, "Unexpected error: "+err.Error(), token)
	} else {
		c := tokens[token]
		file, err := getPath(fname, c)
		if err != nil{
			clientError(conn, "Path error: "+err.Error(), token)
		}else{		
			err = utils.WriteFile(file, buffer)
			if err != nil {
				clientError(conn, "Write error: "+err.Error(), token)
			} else {
				utils.Log("Successfully wrote for: ", c.Name)
				conn.Close()
			}
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
		val, err := utils.JSONValue(jobj, "TYPE")
		if err != nil {
			clientError(conn, "Invalid request type "+string(data)+" error "+err.Error(), "")
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
		clientError(conn, "Error retrieving JSON value from request", "")
	} else {
		num, err := getProjectNum(uname, pword, project) 
		if err == nil {
			client := client.NewClient(uname, project, num, mode)
			token := getToken(client)
			err := utils.CreateUserProject(client)
			if err != nil {
				clientError(conn, "Error creating project: "+err.Error(), token)
			} else {
				conn.Write([]byte("TOKEN:" + token))
			}
		} else {
			clientError(conn, "Invalid username, password or project "+uname+" "+pword + " " +project, "")
		}
	}
}

func handleSend(jobj map[string]interface{}, conn net.Conn) {
	token, errt := utils.JSONValue(jobj, "TOKEN")
	fname, errf := utils.JSONValue(jobj, "FILENAME")
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
	token, errt := utils.JSONValue(jobj, "TOKEN")
	if errt != nil {
		clientError(conn, "Error retrieving JSON value from request", "")
	} else {
		if tokens[token] != nil {
			c := tokens[token]
			err := utils.ZipProject(c)
			if err == nil{
				err = utils.Remove(c.Project+utils.SEP+c.Name)
			}
			conn.Write([]byte("ACCEPT"))
			if err != nil {
				clientError(conn, "Error creating zip file "+token, token)
			} else {
				conn.Close()
			}
		} else {
			clientError(conn, "Invalid token "+token, "")
		}
	}
}

func handleLogout(jobj map[string]interface{}, conn net.Conn) {
	token, err := utils.JSONValue(jobj, "TOKEN")
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
	utils.Log(msg)
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
