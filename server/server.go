package server

import (
	"bytes"
	"errors"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
	"io"
	"net"
	"labix.org/v2/mgo/bson"
)


/*
Determines whether a file send request is valid and reads a file if it is.
*/
func ProcessFile(subId bson.ObjectId, jobj map[string]interface{}, conn net.Conn) (err error) {
	fname, err := utils.JSONValue(jobj, "FILENAME")
	if err == nil {
		conn.Write([]byte("OK"))
		buffer, err := utils.ReadFile(conn, []byte("EOF")) 
		if err == nil {
			_, err = db.AddFile(subId, fname, buffer.Bytes())
			if err == nil{
				conn.Write([]byte("OK"))
			}
		}
	}
	return err
}


/*
Manages an incoming connection request.
*/
func ConnHandler(conn net.Conn) {
	jobj, err := utils.ReadJSON(conn)
	if err == nil {
		subId, err := Login(jobj, conn)
		for err == nil{
			jobj, err = utils.ReadJSON(conn)
			if err == nil{
				req, err := utils.JSONValue(jobj, "TYPE")
				if req == "SEND" {
					err = ProcessFile(subId, jobj, conn)
				} else if req == "LOGOUT" {
					break
				} else if req == "TESTS" {
					err = ProcessTests(jobj, conn)
				} else if err == nil{
					err = errors.New("Unknown request: " + req)
				}
			}
		}
	}
	EndSession(conn, err)
}

/*
Determines whether a login request is valid and delivers this 
result to the client 
*/
func Login(jobj map[string]interface{}, conn net.Conn) (subId bson.ObjectId, err error) {
	c, err := createClient(jobj)
	if err == nil {
		subId, err = db.CreateSubmission(c.project, c.username, c.format)
		if err == nil{
			conn.Write([]byte("OK"))
		}
	}
	return subId, err
}

/*
Receives new project tests and stores them in the database.
*/
func ProcessTests(jobj map[string]interface{}, conn net.Conn) (err error) {
	project, err := utils.JSONValue(jobj, "PROJECT")
	if err == nil {
		conn.Write([]byte("OK"))
		buffer := new(bytes.Buffer)
		p := make([]byte, 2048)
		bytesRead, err := conn.Read(p)
		for err == nil {
			buffer.Write(p[:bytesRead])
			bytesRead, err = conn.Read(p)
		}
		if err == io.EOF {
			err = db.AddTests(project, buffer.Bytes())
		}
	}
	return err
}



/*
Handles an error by logging it as well as reporting it to the connected
user if possible.
*/
func EndSession(conn net.Conn, err error) {
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		utils.Log(msg)
	} else{
		msg = "OK"
	}
	conn.Write([]byte(msg))
	conn.Close()
}

type Client struct{
	username string
	project string
	format string
}

func createClient(jobj map[string]interface{}) (c *Client, err error) {
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
	user, err := db.ReadUser(uname)
	if err == nil {
		if utils.Validate(user.Password, user.Salt, pword) {
			c = &Client{uname, project, format}
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
