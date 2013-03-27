package client

const ONSAVE = "ONSAVE"
const ONSTOP = "ONSTOP"

type Client struct {
	Name       string
	Project    string
	Token string
	Format       string
	SubNum int
	
}

func NewClient(name, project, token, mode string) *Client {
	return &Client{name, project, token, mode, -1}
}

