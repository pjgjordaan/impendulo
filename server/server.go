package server

import (
	"bytes"
	"container/list"
	"fmt"
	"intlola/client"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"archive/zip"
)

func Write(c *client.Client, data *bytes.Buffer) error {
	Log("Writing to: ", c.Project+"/"+c.File)
	out := c.Project
	os.Mkdir(out, 0777)
	if !c.Zip{
		out = out+string(os.PathSeparator)+c.Name
		os.Mkdir(out, 0777)
	}
	out = out + string(os.PathSeparator) + c.File
	err := ioutil.WriteFile(out, data.Bytes(), 0666)
	return err
}
func Log(v ...interface{}) {
	fmt.Println(v...)
}

func ZipFiles(c *client.Client){
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	finfos, err := ioutil.ReadDir(c.Project+string(os.PathSeparator)+c.Name)
	if err != nil{
		Log(err)
	} else{
	for _, file := range finfos {
		if !file.IsDir(){
			f, err := w.Create(file.Name())
			if err != nil {
				Log(err)
				break
			}
			contents, err := ioutil.ReadFile(file.Name())
			if err != nil{
				Log(err)
				break
			}
			_, err = f.Write(contents)
			if err != nil {
				Log(err)
				break
			}
		}
	}
	}
	err = w.Close()
	if err != nil {
		Log(err)
	} else{
		c.Zip = true
		Write(c, buf)
	}
}

func Remove(c *client.Client) {
	c.Conn.Close()
	for entry := clientList.Front(); entry != nil; entry = entry.Next() {
		cur := entry.Value.(client.Client)
		if c.Equal(&cur) {
			Log("Removed: ", c.Name)
			clientList.Remove(entry)
			break
		}
	}
}
func FileReader(c *client.Client) {
	buffer := new(bytes.Buffer)
	p := make([]byte, 2048)
	bytesRead, err := c.Conn.Read(p)
	for err == nil {
		buffer.Write(p[:bytesRead])
		bytesRead, err = c.Conn.Read(p)
	}
	if err != io.EOF {
		Log("Client ", c.Name, " resulted in unexpected error: ", err)
	}
	Log("Reader stopped for ", c.Name)
	err = Write(c, buffer)
	if err != nil {
		Log(c.Name, " had write error: ", err)
	} else {
		Log("Successfully wrote for: ", c.Name)
	}
	Remove(c)
}

var ARGS_LEN int = 4

func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		conn.Close()
		Log("Client connection error: ", error)
	}
	req := strings.TrimSpace(string(buffer[:bytesRead]))
	args := strings.Split(req, ":")
	if len(args) != ARGS_LEN {
		conn.Close()
		Log("Invalid number of arguments in connection request: " + req)
	} else{
		name := args[1]
		project := args[2]
		fname := args[3]
		c := client.NewClient(name, project, fname, conn)
		Log("Connected to new client ", name, " sending ", fname, " from project ", project, "  on ", conn.RemoteAddr())
		if args[0] == "SEND"{	
			c.Conn.Write([]byte("ACCEPT"))
			clientList.PushBack(*c)
			FileReader(c)
		} else if args[0] == "DONE"{
			ZipFiles(c)
		} else {
			Log("Invalid connection request: " + req)
			c.Conn.Close()
		}
	}
}


var clientList *list.List = list.New()

func Run(address string, port string) {
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
				Log("Waiting for connections")
				connection, error := netListen.Accept()
				if error != nil {
					Log("Client error: ", error)
				} else {
					go ConnHandler(connection)
				}
			}
		}
	}
}
