package client

import( "net")

type Client struct {
	Token byte
	Conn net.Conn
	File string
}

func NewClient(tok byte, con net.Conn, fname string)(*Client){
	return &Client{tok, con, fname}
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
	c.Conn.Close()
}