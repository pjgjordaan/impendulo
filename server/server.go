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
	"labix.org/v2/mgo/bson"
)

/*
Reads files from a client's TCP connection and stores them.
*/
func fileReader(conn net.Conn, token string, fname string) (err error) {
	buffer := new(bytes.Buffer)
	p := make([]byte, 2048)
	receiving := true
	eof := []byte("EOF") 
	for receiving {
		bytesRead, err := conn.Read(p)
		read := p[:bytesRead]
		if bytes.HasSuffix(read, eof) || err != nil{
			read = read[:len(read)-len(eof)]
			receiving = false
		} 
		if err == nil || err == io.EOF{
			buffer.Write(read)
		}
	}
	if err == io.EOF || err == nil {
		err = nil
		if c, ok := getClient(token); ok {
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
	var msg string
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
					msg, err = handleLogin(jobj)
				} else if val == "SEND" {
					msg, err = handleSend(jobj, conn)
				} else if val == "LOGOUT" {
					msg, err = handleLogout(jobj)
				} else if val == "TESTS" {
					msg, err = handleTests(jobj, conn)
				} else {
					err = errors.New("Unknown request: " + val)
				}
			}
		}
	}
	endSession(conn, msg, err)
}

/*
Determines whether a login request is valid and delivers this 
result to the client 
*/
func handleLogin(jobj map[string]interface{}) (msg string, err error) {
	c, err := createClient(jobj)
	if err == nil {
		num, err := db.CreateSubmission(c)
		c.SubNum = num
		if err == nil {
			msg = "TOKEN:" + c.Token
		}
	}
	return msg, err
}

/*
Receives new project tests and stores them in the database.
*/
func handleTests(jobj map[string]interface{}, conn net.Conn) (msg string, err error) {
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
			err = db.AddTests(project, buffer.Bytes())
			if err == nil{
				msg = "Successfully received tests."
			}
		}
	}
	return msg, err
}

/*
Determines whether a file send request is valid and reads a file if it is.
*/
func handleSend(jobj map[string]interface{}, conn net.Conn) (msg string, err error) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err == nil {
		fname, err := utils.JSONValue(jobj, "FILENAME")
		if err == nil {
			c, ok := getClient(token)
			if  ok{
				conn.Write([]byte("ACCEPT"))
				err = fileReader(conn, c, fname)
				if err == nil{
					msg = "Successfully received file."
					fdata := bson.M{"project" : c.Project, "user" : c.Name, "number" : c.SubNum, "format" : c.Format, "file": fname}  
					go utils.ProcessFile(fdata) 
				}
			} else {
				err = errors.New("Invalid token")
			}
		}
	}
	return msg, err
}

/*
Ends a user's session.
*/
func handleLogout(jobj map[string]interface{}) (msg string, err error) {
	token, err := utils.JSONValue(jobj, "TOKEN")
	if err == nil {
		_, ok = getClient(token)
		if ok {
			msg = "ACCEPT"
		} else {
			err = errors.New("Invalid token")
		}
	}
	return msg, err
}

/*
Handles an error by logging it as well as reporting it to the connected
user if possible.
*/
func endSession(conn net.Conn, msg string, err error) {
	if err != nil {
		msg = "Error encountered: " + err.Error()
		utils.Log(msg)
	}
	if conn != nil {
		conn.Write([]byte(msg))
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
			addClient(c)
		} else {
			err = errors.New("Invalid username or password")
		}
	}
	return c, err
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
			reads := make(chan *readOp)
			writes := make(chan *writeOp)
			go tokenHandler()
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
