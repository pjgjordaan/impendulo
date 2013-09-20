//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
