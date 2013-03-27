package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"io"
	"net"
)

var tokens map[string]*client.Client

/*
Initialises the tokens available for use.
*/
func init() {
	tokens = make(map[string]*client.Client)
}

/*
Reads files from a client's TCP connection and stores them.
*/
func fileReader(conn net.Conn, token string, fname string) (err error) {
	buffer := new(bytes.Buffer)
	p := make([]byte, 2048)
	bytesRead, err := conn.Read(p)
	for err == nil {
		buffer.Write(p[:bytesRead])
		bytesRead, err = conn.Read(p)
	}
	if err == io.EOF {
		err = nil
		if c, ok := tokens[token]; ok {
			err = db.AddFile(c, fname, buffer.Bytes())
		} else {
			err = errors.New("Invalid token: " + token)
		}
	}
	return err
}

/*
Manages an incoming connection request.
*/
func connHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err == nil {
		buffer = buffer[:bytesRead]
		var holder interface{}
		err = json.Unmarshal(buffer, &holder)
		if err == nil {
			jobj := holder.(map[string]interface{})
			val, err := utils.JSONValue(jobj, "TYPE")
			if err == nil {
				if val == "LOGIN" {
					err = handleLogin(jobj, conn)
				} else if val == "SEND" {
					err = handleSend(jobj, conn)
				} else if val == "LOGOUT" {
					err = handleLogout(jobj, conn)
				} else if val == "TESTS" {
					err = handleTests(jobj, conn)
				} else {
					err = errors.New("Unknown request: " + val)
				}
			}
		}
	}
	endSession(conn, err)
}

/*
Determines whether a login request is valid and delivers this 
result to the client 
*/
func handleLogin(jobj map[string]interface{}, conn net.Conn) (err error) {
	c, err := createClient(jobj)
	if err == nil {
		num, err := db.CreateSubmission(c)
		c.SubNum = num
		if err == nil {
			_, err = conn.Write([]byte("TOKEN:" + c.Token))
		}
	}
	return err
}

/*
Receives new project tests and stores them in the database.
*/
func handleTests(jobj map[string]interface{}, conn net.Conn) (err error) {
	project, err := utils.JSONValue(jobj, "PROJECT")
	if err == nil {
		conn.Write([]byte("ACCEPT"))
		buffer := new(bytes.Buffer)
		p := make([]byte, 2048)
		bytesRead, err := conn.Read(p)
		for err == nil {
			buffer.Write(p[:bytesRead])
			bytesRead, err = conn.Read(p)
		}
		if err == io.EOF {
			err = nil
			err = db.AddTests(project, buffer.Bytes())
		}
	}
	return err
}

/*
Determines whether a file send request is valid and reads a file if it is.
*/
func handleSend(jobj map[string]interface{}, conn net.Conn) (err error) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err == nil {
		fname, err := utils.JSONValue(jobj, "FILENAME")
		if err == nil {
			if tokens[token] != nil {
				conn.Write([]byte("ACCEPT"))
				err = fileReader(conn, token, fname)
			} else {
				err = errors.New("Invalid token")
			}
		}
	}
	return err
}

/*
Ends a user's session.
*/
func handleLogout(jobj map[string]interface{}, conn net.Conn) (err error) {
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
	return err
}

/*
Handles an error by logging it as well as reporting it to the connected
user if possible.
*/
func endSession(conn net.Conn, err error) {
	if err != nil {
		errMsg := "Error encountered: " + err.Error()
		utils.Log(errMsg)
		if conn != nil {
			conn.Write([]byte(errMsg))
		}
	}
	if conn != nil {
		conn.Close()
	}
}

func createClient(jobj map[string]interface{}) (c *client.Client, err error) {
	uname, err := utils.JSONValue(jobj, "USERNAME")
	if err != nil {
		return c, err
	}
	pword, err := utils.JSONValue(jobj, "PASSWORD")
	if err != nil {
		return c, err
	}
	project, err := utils.JSONValue(jobj, "PROJECT")
	if err != nil {
		return c, err
	}
	format, err := utils.JSONValue(jobj, "FORMAT")
	if err != nil {
		return c, err
	}
	c, err = setupClient(uname, pword, project, format)
	return c, err
}

/*
Authenticates and assigns a token to a newly connected  client.
*/
func setupClient(uname, passwd, project, format string) (c *client.Client, err error) {
	user, err := db.ReadUser(uname)
	if err == nil {
		if utils.Validate(user.Password, user.Salt, passwd) {
			tok := genToken()
			c = client.NewClient(uname, project, tok, format)
			tokens[tok] = c
		} else {
			err = errors.New("Invalid username or password")
		}
	}
	return c, err
}

/*
Generates a new token.
*/
func genToken() (tok string) {
	for {
		tok = utils.GenString(32)
		if tokens[tok] == nil {
			break
		}
	}
	return tok
}

/*
Listens for new connections and creates a new goroutine for each connection.
*/
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
					go connHandler(conn)
				}
			}
		}
	}
}
