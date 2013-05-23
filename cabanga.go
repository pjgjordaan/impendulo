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

var fport, tport, ufile, mode string

//init is used here to setup flags. 
func init() {
	flag.StringVar(&fport, "fp", "9000", "Specify the port to listen on for files.")
	flag.StringVar(&tport, "tp", "8000", "Specify the port to listen on for tests.")
	flag.StringVar(&ufile, "u", "", "Specify a file with new users.")
	flag.StringVar(&mode, "m", "s", "Specify a mode to run in.")
}

func main() {
	flag.Parse()
	if mode == "u" {
		err := AddUsers()
		if err != nil {
			util.Log(err)
		}
	} else if mode == "s" {
		Run()
	} else {
		log.Fatal(fmt.Errorf("Unknown running mode %q", mode))
	}
}

//AddUsers adds users from a text file to the database.
//Any errors encountered are returned.
func AddUsers() error {
	users, err := util.ReadUsers(ufile)
	if err != nil {
		return err
	}
	db.Setup(db.DEFAULT_CONN)
	return db.AddUsers(users...)
}

//RunServer starts an instance of our tcp snapshot server on the given port.
//A seperate routine is launched which processes the snapshots.
func Run() {
	util.Log("Starting server on port ", fport)
	runtime.GOMAXPROCS(runtime.NumCPU())
	db.Setup(db.DEFAULT_CONN)
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go processing.Serve(subChan, fileChan)
	go server.Run(tport, new(server.TestSpawner))
	server.Run(fport, &server.SubmissionSpawner{subChan, fileChan})
}
