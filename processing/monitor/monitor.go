//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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

func init() {
	busy, done = make(chan bson.ObjectId), make(chan bson.ObjectId)

}

func SetStore(fname string) {
	storeName = fname
}

func Busy(subId bson.ObjectId) {
	busy <- subId
}

func Done(subId bson.ObjectId) {
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
		case sig := <-quit:
			util.Log(sig)
			err := saveActive(active)
			if err != nil {
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
func saveActive(active map[bson.ObjectId]bool) error {
	err := util.SaveMap(active, filepath.Join(util.BaseDir(), storeName))
	if err != nil {
		return err
	}
	return nil
}
