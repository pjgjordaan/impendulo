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
)

var tokens map[string]*client.Client

const STRLEN = 80

func init() {
	tokens = make(map[string]*client.Client)
	used, err := db.GetAll("token")
	if handleError(nil, "", "Error retrieving submission tokens ", err){
		dummy := client.NewClient("", "", "", "")
		for _, tok := range used{
			tokens[tok] = dummy
		}
	}
	
}

func getPath(fname string, c *client.Client) (file string, err error) {
	if c.Mode == client.ONSAVE {
		file = c.Project + utils.SEP + c.Name + utils.SEP + fname
	} else if c.Mode == client.ONSTOP {
		file = c.Project + utils.SEP + c.Project+ "_" + c.Name + "_" + c.Token + ".zip"
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
		err = db.AddFile(c, fname, buffer.Bytes())
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
			} else {
				err = errors.New("Unknown request: "+val)
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
		token := createClient(uname, project, mode)
		err := db.CreateProject(tokens[token])
		if err == nil{
			_, err = conn.Write([]byte("TOKEN:" + token))
		}
		handleError(conn, "","Login IO Error - username: "+uname+" password: "+pword, err)
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
	if conn != nil{
		conn.Write([]byte(errMsg))
		conn.Close()
	}
	if token != "" {
		delete(tokens, token)
	}
	return false
}

func genToken() string {
	b := make([]byte, STRLEN)
	rand.Read(b)
	en := base64.StdEncoding
	d := make([]byte, en.EncodedLen(len(b)))
	en.Encode(d, b)
	return string(d)
}

func createClient(uname, project, mode string) (string) {
	tok := genToken()
	for tokens[tok] != nil {
		tok = genToken()
	}
	c := client.NewClient(uname, project, tok, mode)
	tokens[tok] = c
	return tok
}

func Run(address string, port string) {
	service := address + ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		utils.Log("Error: Could not resolve address ", err)
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
