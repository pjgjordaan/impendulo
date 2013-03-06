package server

import( 
"fmt"
"net"
"container/list"
"strings"
"io"
"io/ioutil"
"intlola/client"
"bytes"
"os"
)


func Write(project string, fname string, data *bytes.Buffer) error{
	Log("Writing to: ", project+"/"+fname)
	os.Mkdir(project, 0666)
	out := project+os.PathSeparator+fname 
	err := ioutil.WriteFile(out, data.Bytes(), 0666)
	return err
}
func Log(v ...interface{}) {
	fmt.Println(v...)
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
	if err != io.EOF{
		Log("Client ", c.Name, " resulted in unexpected error: ", err) 
	}
	Log("Reader stopped for ", c.Name)
	err = Write(c.Name+c.File, buffer)
	if err != nil{
		Log(c.Name, " had write error: ", err)
	} else{
		Log("Successfully wrote for: " , c.Name)
	}
	Remove(c)
}

var NUMARGS int = 4 
func ConnHandler(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, error := conn.Read(buffer)
	if error != nil {
		Log("Client connection error: ", error)
		return
	}
	req := strings.TrimSpace(string(buffer[:bytesRead]))
	args := strings.Split(req,":")
	if args[0] !=  "CONNECT" || len(args) != NUMARGS{
		Log("Invalid connection request: "+req)
		return
	}
	name := args[1]
	project := args[2]
	fname := args[3] 
	newClient := client.NewClient(name, project, fname, conn)
	Log("Connected to new client ", name, " sending ", fname, " from project ", project ,"  on ", conn.RemoteAddr())
	conn.Write([]byte("ACCEPT"))
	clientList.PushBack(*newClient)
	FileReader(newClient)
}

var clientList *list.List = list.New()

func Run(address string, port string){
	Log("Server Started")
	service := address+":"+port
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

