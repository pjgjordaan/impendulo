package main

import (
	"flag"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/processing"
	"github.com/godfried/cabanga/server"
	"github.com/godfried/cabanga/project"
	"github.com/godfried/cabanga/user"
	"github.com/godfried/cabanga/util"
	"github.com/godfried/cabanga/config"
	"runtime"
)

//Flag variables for setting ports to listen on, users file to process and the mode to run in.
	var FilePort, TestPort, UsersFile, ConfigFile string

func init() {
	flag.StringVar(&FilePort, "fp", "8010", "Specify the port to listen on for files.")
//	flag.StringVar(&TestPort, "tp", "8011", "Specify the port to listen on for tests.")
	flag.StringVar(&UsersFile, "u", "", "Specify a file with new users.")
	flag.StringVar(&ConfigFile, "c", "config.txt", "Specify a configuration file.")
}

func main() {
	flag.Parse()
	err := config.LoadConfigs(ConfigFile)
	if err != nil{
		panic(err)
	}
	if UsersFile != "" {
		err := AddUsers()
		if err != nil {
			util.Log(err)
		}
	}
	Run()
}

//AddUsers adds users from a text file to the database.
func AddUsers() error {
	users, err := user.ReadUsers(UsersFile)
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
	fileChan := make(chan *project.File)
	subChan := make(chan *project.Submission)
	go processing.Serve(subChan, fileChan)
	//go server.Run(TestPort, new(server.TestSpawner))
	server.Run(FilePort, &server.SubmissionSpawner{subChan, fileChan})
}
