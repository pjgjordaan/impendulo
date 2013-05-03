package server

import (
	"errors"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/proc"
	"github.com/disco-volante/intlola/sub"
	"github.com/disco-volante/intlola/user"
	"github.com/disco-volante/intlola/utils"
	"labix.org/v2/mgo/bson"
	"net"
)

type Client struct {
	username string
	project  string
	mode     string
	lang     string
}

/*
Manage incoming connection request.
*/
func ConnHandler(conn net.Conn, fileChan chan bson.ObjectId) {
	jobj, err := utils.ReadJSON(conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	subId, err := Login(jobj, conn)
	if err != nil {
		EndSession(conn, err)
		return
	}
	utils.Log("Created submission: ", subId)
	receiving := true
	for receiving && err == nil {
		jobj, err = utils.ReadJSON(conn)
		if err != nil {
			utils.Log("JSON error: ", err)
			EndSession(conn, err)
			return
		}
		req, err := utils.GetString(jobj, REQ)
		if err != nil {
			utils.Log("JSON error: ", err)
			EndSession(conn, err)
			return
		}
		if req == SEND {
			delete(jobj, REQ)
			err = ProcessFile(subId, jobj, conn, fileChan)
		} else if req == LOGOUT {
			receiving = false
			utils.Log("Completed submission: ", subId)
		} else {
			err = errors.New("Unknown request: " + req)
		}
	}
	EndSession(conn, err)
}

/*
Reads file data from connection and sends data to be processed.
*/
func ProcessFile(subId bson.ObjectId, finfo map[string]interface{}, conn net.Conn, fileChan chan bson.ObjectId) error {
	conn.Write([]byte(OK))
	buffer, err := utils.ReadData(conn, []byte(EOF))
	if err != nil {
		utils.Log("Conn read error: ", err)
		return err
	}
	utils.Log("Read file: ", finfo)
	conn.Write([]byte(OK))
	f := sub.NewFile(subId, finfo, buffer.Bytes())
	err = db.AddOne(db.FILES, f)
	if err != nil {
		utils.Log("DB error: ", err)
		return err
	}
	utils.Log("Saved file: ", f.Id)
	fileChan <- f.Id
	return nil
}

/*
Creates a new submission if the login request is valid.
*/
func Login(jobj map[string]interface{}, conn net.Conn) (subId bson.ObjectId, err error) {
	c, err := createClient(jobj)
	if err != nil {
		utils.Log("Login error: ", err)
		return subId, err
	}
	s := sub.NewSubmission(c.project, c.username, c.mode, c.lang)
	err = db.AddOne(db.SUBMISSIONS, s)
	if err != nil {
		utils.Log("DB error: ", err)
		return subId, err
	}
	conn.Write([]byte(OK))
	subId = s.Id
	return subId, err
}

/*
Ends a client session and reports any errors to the client. 
*/
func EndSession(conn net.Conn, err error) {
	var msg string
	if err != nil {
		msg = "ERROR: " + err.Error()
		utils.Log("Sending error: ", msg)
	} else {
		msg = OK
	}
	conn.Write([]byte(msg))
	conn.Close()
}

/*
Reads client values from a json object.
Determines whether client has neccesary privileges for submission and is using correct password 
*/
func createClient(jobj map[string]interface{}) (c *Client, err error) {
	uname, err := utils.GetString(jobj, UNAME)
	if err != nil {
		return c, err
	}
	pword, err := utils.GetString(jobj, PWORD)
	if err != nil {
		return c, err
	}
	project, err := utils.GetString(jobj, PROJECT)
	if err != nil {
		return c, err
	}
	mode, err := utils.GetString(jobj, MODE)
	if err != nil {
		return c, err
	}
	lang, err := utils.GetString(jobj, LANG)
	if err != nil {
		return c, err
	}
	umap, err := db.GetById(db.USERS, uname)
	if err == nil {
		usr := user.ReadUser(umap)
		if usr.CheckSubmit(mode) && utils.Validate(usr.Password, usr.Salt, pword) {
			c = &Client{uname, project, mode, lang}
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
	fileChan := make(chan bson.ObjectId)
	go proc.Serve(fileChan)
	service := address + ":" + port
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		utils.Log("Error: Could not resolve address ", err)
	} else {
		netListen, err := net.Listen(tcpAddr.Network(), tcpAddr.String())
		if err != nil {
			utils.Log("Listening error: ", err)
			return
		}
		defer netListen.Close()
		for {
			conn, err := netListen.Accept()
			if err != nil {
				utils.Log("Client error: ", err)
			} else {
				go ConnHandler(conn, fileChan)
			}
		}
	}
}

const (
	OK      = "ok"
	EOF     = "eof"
	SEND    = "send"
	LOGIN   = "begin"
	LOGOUT  = "end"
	REQ     = "req"
	UNAME   = "uname"
	PWORD   = "pword"
	PROJECT = "project"
	MODE    = "mode"
	LANG    = "lang"
)
