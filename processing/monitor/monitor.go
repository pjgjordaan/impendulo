package monitor

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"os/signal"
	"path/filepath"
)
const INCOMPLETE = "incomplete.gob"

var busy, done chan bson.ObjectId
var storeName string = INCOMPLETE

func init(){
	busy, done = make(chan bson.ObjectId), make(chan bson.ObjectId)
	
}

func SetStore(fname string){
	storeName = fname
}

func Busy(subId bson.ObjectId){
	busy <- subId
}

func Done(subId bson.ObjectId){
	done <- subId
}

//Monitor listens for new submissions and adds them to the map of active processes. 
//It also listens for completed submissions and removes them from the active process map.
//Finally it detects Kill and Interrupt signals, saving the active processes if they are detected.
func Listen() {	
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Kill, os.Interrupt)
	active := GetStored()
	for {
		select {
		case id := <-busy:
			active[id] = true
		case id := <-done:
			delete(active, id)
		case <-quit:
			err := saveActive(active)
			if err != nil{
				util.Log(err)
			}
			os.Exit(0)
			return
		}
	}
}

//getStored retrieves incompletely processed submissions from the filesystem.
func GetStored() map[bson.ObjectId]bool {
	stored, err := util.LoadMap(filepath.Join(util.BaseDir(), storeName))
	if err != nil {
		util.Log(err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}

//saveActive saves active submissions to the filesystem.
func saveActive(active map[bson.ObjectId]bool)error {
	err := util.SaveMap(active, filepath.Join(util.BaseDir(), storeName))
	if err != nil {
		return err
	}
	return nil
}
