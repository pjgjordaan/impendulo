package client

import( "net")

type Client struct {
	Name string
	Conn net.Conn
	File string
}

func NewClient(name string, con net.Conn, fname string)(*Client){
	return &Client{name, con, fname}
}

func (c *Client) Equal(other *Client) bool {
	if c.Name == other.Name {
		if c.Conn == other.Conn {
			return true
		}
	}
	return false
}
