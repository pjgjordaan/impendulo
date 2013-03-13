package client

const ONSAVE = "ONSAVE"
const ONSTOP = "ONSTOP"

type Client struct {
	Name       string
	Project    string
	ProjectNum int
	Mode       string
}

func NewClient(name string, project string, num int, mode string) *Client {
	return &Client{name, project, num, mode}
}

type ClientData struct {
	Projects map[string]int
	Password string
}
