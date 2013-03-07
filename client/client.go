package client

import (
	"strings"
	"net"
)

type Client struct {
	Name    string
	Project string
	File    string
	Conn    net.Conn
	Zip bool
}

func NewClient(name string, project string, fname string, con net.Conn) *Client {
	var c *Client
	if strings.HasSuffix(fname, "zip"){
		c = &Client{name, project, fname, con, true} 
	} else {
		c = &Client{name, project, fname, con, false}
	}
	return c
}

func (c *Client) Equal(other *Client) bool {
	if c.Name == other.Name && c.Conn == other.Conn && c.Project == other.Project && c.File == other.File {
		return true
	}
	return false
}
