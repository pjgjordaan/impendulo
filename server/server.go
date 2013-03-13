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
		c := tokens[token]
		file, err := getPath(fname, c)
		if err == nil {
			err = utils.WriteFile(file, buffer)
		}
	}
	if handleError(conn, token, err) {
		conn.Close()
	}
}

func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if handleError(conn, "", err) {
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
	handleError(conn, "", err)
}

func handleLogin(jobj map[string]interface{}, conn net.Conn) {
	uname, erru := utils.JSONValue(jobj, "USERNAME")
	pword, errw := utils.JSONValue(jobj, "PASSWORD")
	project, errp := utils.JSONValue(jobj, "PROJECT")
	mode, errm := utils.JSONValue(jobj, "MODE")
	if handleError(conn, "", erru, errw, errp, errm) {
		num, err := getProjectNum(uname, pword, project)
		if err == nil {
			c := client.NewClient(uname, project, num, mode)
			token := getToken(c)
			err := utils.MkDir(c.Project + utils.SEP + c.Name)
			if err == nil {
				conn.Write([]byte("TOKEN:" + token))
			}
		}
		handleError(conn, "", err)
	}
}

func handleSend(jobj map[string]interface{}, conn net.Conn) {
	token, errt := utils.JSONValue(jobj, "TOKEN")
	fname, errf := utils.JSONValue(jobj, "FILENAME")
	if handleError(conn, "", errt, errf) {
		if tokens[token] != nil {
			conn.Write([]byte("ACCEPT"))
			FileReader(conn, token, fname)
		} else {
			handleError(conn, "", errors.New("Invalid token"))
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
	handleError(conn, "", err)
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
	handleError(conn, "", err)
}

func handleError(conn net.Conn, token string, errs ...error) bool {
	if errs == nil {
		return true
	}
	utils.Log(token, errs)
	errMsg := "Error encountered: "
	for _, err := range errs {
		errMsg += err.Error()
	}
	conn.Write([]byte(errMsg))
	conn.Close()
	if token != "" {
		c := tokens[token]
		utils.Remove(c.Project + utils.SEP + c.Name)
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
