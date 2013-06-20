package main

import (
	"flag"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/server/web"	
	"github.com/godfried/impendulo/server"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/config"
	"runtime"
)

//Flag variables for setting ports to listen on, users file to process and the mode to run in.
var FilePort, TestPort, UsersFile, ConfigFile string

func init() {
	flag.StringVar(&FilePort, "fp", "8010", "Specify the port to listen on for files.")
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
	go server.Run(FilePort, &server.SubmissionSpawner{})
	go web.Run()
	processing.Serve()
	
}
