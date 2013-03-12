package client

type Client struct {
	Name       string
	Project    string
	ProjectNum int
	Mode       string
}

const ONSAVE = "ONSAVE"
const ONSTOP = "ONSTOP"

func NewClient(name string, project string, num int, mode string) *Client {
	return &Client{name, project, num, mode}
}

func (c *Client) Equal(other *Client) bool {
	if c.Name == other.Name && c.Project == other.Project {
		return true
	}
	return false
}

type ClientData struct {
	Projects map[string]int
	Password string
}
