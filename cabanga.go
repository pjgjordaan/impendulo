package main

import (
	"flag"
	"fmt"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/processing"
	"github.com/godfried/cabanga/server"
	"github.com/godfried/cabanga/submission"
	"github.com/godfried/cabanga/util"
	"log"
	"runtime"
)

var port, users, mode string

//init is used here to setup flags. 
func init() {
	flag.StringVar(&port, "p", "9000", "Specify the port to listen on.")
	flag.StringVar(&users, "u", "", "Specify a file with new users.")
	flag.StringVar(&mode, "m", "s", "Specify a mode to run in.")
}

func main() {
	flag.Parse()
	if mode == "u" {
		err := AddUsers(users)
		if err != nil {
			util.Log(err)
		}
	} else if mode == "s" {
		RunServer(port)
	} else {
		log.Fatal(fmt.Errorf("Unknown running mode %q", mode))
	}
}

//AddUsers adds users from a text file to the database.
//Any errors encountered are returned.
func AddUsers(fname string) error {
	users, err := util.ReadUsers(fname)
	if err != nil {
		return err
	}
	return db.AddUsers(users...)
}

//RunServer starts an instance of our tcp snapshot server on the given port.
//A seperate routine is launched which processes the snapshots.
func RunServer(port string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	util.Log("Starting server on port ", port)
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go processing.Serve(subChan, fileChan)
	server.Run(port, subChan, fileChan)
}
