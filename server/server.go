package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"io"
	"net"
	"strconv"
)

var tokens map[string]*client.Client

const STRLEN = 80

func init() {
	tokens = make(map[string]*client.Client)
}

func getPath(fname string, c *client.Client) (file string, err error) {
	if c.Mode == client.ONSAVE {
		file = c.Project + utils.SEP + c.Name + utils.SEP + fname
	} else if c.Mode == client.ONSTOP {
		file = c.Project + utils.SEP + c.Name + "_" + c.Project + strconv.Itoa(c.ProjectNum) + ".zip"
	} else {
		err = errors.New("Unknown send mode: " + c.Mode)
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
	if err == io.EOF {
		err = nil
		c := tokens[token]
		file, err := getPath(fname, c)
		if err == nil {
			err = utils.WriteFile(file, buffer)
		}
	}
	if handleError(conn, token, "File read error - ",  err) {
		conn.Close()
	}
}

func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if handleError(conn, "", "Connection error - ", err) {
		handleRequest(buffer[:bytesRead], conn)
	}
}

func handleRequest(data []byte, conn net.Conn) {
	var holder interface{}
	err := json.Unmarshal(data, &holder)
	if err == nil {
		jobj := holder.(map[string]interface{})
		val, err := utils.JSONValue(jobj, "TYPE")
		if err == nil {
			if val == "LOGIN" {
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
	handleError(conn, "", "Request error - ", err)
}

func handleLogin(jobj map[string]interface{}, conn net.Conn) {
	uname, erru := utils.JSONValue(jobj, "USERNAME")
	pword, errw := utils.JSONValue(jobj, "PASSWORD")
	project, errp := utils.JSONValue(jobj, "PROJECT")
	mode, errm := utils.JSONValue(jobj, "MODE")
	if handleError(conn, "","Login JSON Error - ", erru, errw, errp, errm) {
		num, err := getProjectNum(uname, pword, project)
		if err == nil {
			c := client.NewClient(uname, project, num, mode)
			token := getToken(c)
			err := utils.MkDir(c.Project + utils.SEP + c.Name)
			if err == nil {
				_, err = conn.Write([]byte("TOKEN:" + token))
			}
		}
		handleError(conn, "","Login IO Error - ", err)
	}
}

func handleSend(jobj map[string]interface{}, conn net.Conn) {
	token, errt := utils.JSONValue(jobj, "TOKEN")
	fname, errf := utils.JSONValue(jobj, "FILENAME")
	if handleError(conn, "","Sending JSON error - ", errt, errf) {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			FileReader(conn, token, fname)
		} else {
			handleError(conn, "", "Sending Validation error - ", errors.New("Invalid token"))
		}
	}
}

func handleZip(jobj map[string]interface{}, conn net.Conn) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err == nil {
		if tokens[token] != nil {
			c := tokens[token]
			err = utils.ZipProject(c)
			if err == nil {
				err = utils.Remove(c.Project + utils.SEP + c.Name)
			}
			conn.Write([]byte("ACCEPT"))
			if err == nil {
				conn.Close()
			}
		} else {
			err = errors.New("Invalid token")
		}
	}
	handleError(conn, "", "Zip error - ",  err)
}

func handleLogout(jobj map[string]interface{}, conn net.Conn) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err == nil {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			conn.Close()
			delete(tokens, token)
		} else {
			err = errors.New("Invalid token")
		}
	}
	handleError(conn, "", "Logout error - ", err)
}

func handleError(conn net.Conn, token string, msg string, errs ...error) bool {
	errMsg := "Error encountered: " + msg
	size := len(errMsg)
	for _, err := range errs {
		if err != nil{
			errMsg += err.Error()
		}
	}
	if size == len(errMsg){
		return true
	}
	utils.Log(token, errMsg)
	conn.Write([]byte(errMsg))
	conn.Close()
	if token != "" {
		c := tokens[token]
		if c != nil{
			utils.Remove(c.Project + utils.SEP + c.Name)
		}
		delete(tokens, token)
	}
	return false
}

func getProjectNum(uname, pword, project string) (int, error) {
	info, err := db.Read(uname)
	num := -1
	if err == nil {
		if info.Password == pword {
			info.Projects[project]++
			num = info.Projects[project]
			err = db.Add(uname, info)
		} else {
			err = errors.New("Invalid password: " + pword)
		}
	}
	return num, err
}

func genToken() string {
	b := make([]byte, STRLEN)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}

func getToken(c *client.Client) string {
	tok := genToken()
	for tokens[tok] != nil {
		tok = genToken()
	}
	tokens[tok] = c
	return tok
}

func Run(address string, port string) {
	service := address + ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		utils.Log("Error: Could not resolve address", err)
	} else {
		netListen, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if err != nil {
			utils.Log(err)
		} else {
			defer netListen.Close()
			for {
				conn, err := netListen.Accept()
				if err != nil {
					utils.Log("Client error: ", err)
				} else {
					go ConnHandler(conn)
				}
			}
		}
	}
}
