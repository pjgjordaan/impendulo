package client

type Client struct {
	Name    string
	Project string
}

func NewClient(name string, project string) *Client {
	return &Client{name, project}
}

func (c *Client) Equal(other *Client) bool {
	if c.Name == other.Name && c.Project == other.Project {
		return true
	}
	return false
}
