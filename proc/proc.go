package proc

import(
	"labix.org/v2/mgo/bson"
	"github.com/disco-volante/intlola/utils"
	"github.com/disco-volante/intlola/db"

)
const MAX = 100

type Request struct{
	FileId bson.ObjectId
	SubId bson.ObjectId
}
func process(r *Request){
	s, err := db.GetFile(r.FileId, r.SubId)
	utils.Log(err, s)
}

func handle(queue chan *Request) {
	for r := range queue {
		process(r)
	}
}

func Serve(clientRequests chan *Request) {
	// Start handlers
	for i := 0; i < MAX; i++ {
		go handle(clientRequests)
	}
	utils.Log("completed")
}