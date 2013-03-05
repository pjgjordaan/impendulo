package client

import( "net")

type Client struct {
	Token byte
	Incoming chan string
	Outgoing chan string
	Conn net.Conn
	Quit chan bool
	File string
	data []byte
	offset int
}

func NewClient(tok byte, out chan string, con net.Conn, fname string)(*Client){
	return &Client{tok, make(chan string), out, con, make(chan bool), fname, make([]byte, 4096), 0}
}

func (c *Client) Equal(other *Client) bool {
	if c.Token == other.Token {
		if c.Conn == other.Conn {
			return true
		}
	}
	return false
}

func (c *Client) Close() {
	c.Quit <- true
	c.Conn.Close()
}
func (c *Client) AddData(data []byte){
	copied := 0
	if c.offset < len(c.data){
		copied = copy(c.data[c.offset:],data)
		c.offset += len(data)
	}
	if(copied < len(data)){
		c.data = append(c.data, data[copied:]...)
		c.offset = len(c.data)
	}
}

func (c *Client) GetData()([]byte){
	return c.data[:c.offset]
}