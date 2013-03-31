package server
import(
	"sync"
	"github.com/disco-volante/intlola/utils"
	"github.com/disco-volante/intlola/client"
	)

var tokens map[string]*client.Client
var tokMutex *sync.Mutex
/*
Initialises the tokens available for use.
*/
func init() {
	tokMutex = &sync.Mutex{}
	tokens = make(map[string]*client.Client)
}


func tokenHandler(){
	for {
		select {
		case read := <-reads:
			read.resp <- state[read.key]
		case write := <-writes:
			state[write.key] = write.val
			write.resp <- true
		}
	}
}


func genToken() (tok string) {
	for {
		tok = utils.GenString(32)
		if !checkToken(tok) {
			break
		}
	}
	return tok
}

func getClient(token string)(c *client.Client, ok bool){
	tokMutex.Lock()
	c, ok = tokens[token]
	tokMutex.Unlock()
	return c, ok
}

func checkToken(token string)(ok bool){
	_, ok = getClient(token)
	return ok
}

func deleteToken(token string){
	tokMutex.Lock()
	delete(tokens, token)
	tokMutex.Unlock()
}

func addClient(c *client.Client){
	tokMutex.Lock()
	tokens[c.Token] = c
	tokMutex.Unlock()
}