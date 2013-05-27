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

//Flag variables for setting ports to listen on, users file to process and the mode to run in.
var FilePort, TestPort, UsersFile, Mode string

func init() {
	flag.StringVar(&FilePort, "fp", "9000", "Specify the port to listen on for files.")
	flag.StringVar(&TestPort, "tp", "8000", "Specify the port to listen on for tests.")
	flag.StringVar(&UsersFile, "u", "", "Specify a file with new users.")
	flag.StringVar(&Mode, "m", "s", "Specify a mode to run in.")
}

func main() {
	flag.Parse()
	if Mode == "u" {
		err := AddUsers()
		if err != nil {
			util.Log(err)
		}
	} else if Mode == "s" {
		Run()
	} else {
		log.Fatal(fmt.Errorf("Unknown running mode %q", Mode))
	}
}

//AddUsers adds users from a text file to the database.
func AddUsers() error {
	users, err := util.ReadUsers(UsersFile)
	if err != nil {
		return err
	}
	db.Setup(db.DEFAULT_CONN)
	return db.AddUsers(users...)
}

//Run starts a routine for processing snapshot submissions as well as a routine for receiving project tests.
//An instance of our tcp snapshot server is then launched. 
func Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	db.Setup(db.DEFAULT_CONN)
	fileChan := make(chan *submission.File)
	subChan := make(chan *submission.Submission)
	go processing.Serve(subChan, fileChan)
	go server.Run(TestPort, new(server.TestSpawner))
	server.Run(FilePort, &server.SubmissionSpawner{subChan, fileChan})
}
