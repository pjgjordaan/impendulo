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
	Name string
	Password string
	Projects map[string]int
}
func NewData(name, pword string)(*ClientData){
	return &ClientData{name, pword, make(map[string]int)}
}
func (c *ClientData) String() (string){
	return "Username: "+c.Name+", Password: "+c.Password
} 
