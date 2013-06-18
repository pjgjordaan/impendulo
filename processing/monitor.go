package processing

import (
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
)


//getStored retrieves incompletely processed submissions from the filesystem.
func getStored(fname string) map[bson.ObjectId]bool {
	stored, err := util.LoadMap(filepath.Join(util.BaseDir(), fname))
	if err != nil {
		util.Log(err)
		stored = make(map[bson.ObjectId]bool)
	}
	return stored
}

//saveActive saves active submissions to the filesystem.
func saveActive(fname string, active map[bson.ObjectId]bool)error {
	err := util.SaveMap(active, filepath.Join(util.BaseDir(), fname))
	if err != nil {
		return err
	}
	return nil
}

//Monitor listens for new submissions and adds them to the map of active processes. 
//It also listens for completed submissions and removes them from the active process map.
//Finally it detects Kill and Interrupt signals, saving the active processes if they are detected.
func Monitor(fname string, busy, done chan bson.ObjectId, quit chan os.Signal) {
	active := getStored(fname)
	for {
		select {
		case id := <-busy:
			active[id] = true
		case id := <-done:
			delete(active, id)
		case <-quit:
			err := saveActive(fname, active)
			if err != nil{
				util.Log(err)
			}
			os.Exit(0)
			return
		}
	}
}
