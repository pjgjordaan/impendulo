package main

import (
	"flag"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/server"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/webserver"
)

//Flag variables for setting ports to listen on, users file to process and the mode to run in.
var Port, UsersFile, ConfigFile string
var Web, Receiver, Processor, ConsoleErrors, ConsoleInfo bool
var MaxProcs, Timeout int

func init() {
	fmt.Sprint()
	flag.IntVar(&Timeout, "t", 10, "Specify the time limit for a tool to run in, in minutes (default 10).")
	flag.IntVar(&MaxProcs, "mp", 10, "Specify the maximum number of goroutines to run when processing submissions (default 10).")
	flag.BoolVar(&ConsoleErrors, "e", false, "Specify whether to log errors to console (default false).")
	flag.BoolVar(&ConsoleInfo, "i", false, "Specify whether to log info to console (default false).")
	flag.BoolVar(&Web, "w", true, "Specify whether to run the webserver (default true).")
	flag.BoolVar(&Receiver, "r", true, "Specify whether to run the Intlola file receiver (default true).")
	flag.BoolVar(&Processor, "s", true, "Specify whether to run the Intlola file processor (default true).")
	flag.StringVar(&Port, "p", "8010", "Specify the port to listen on for files.")
	flag.StringVar(&UsersFile, "u", "", "Specify a file with new users.")
	flag.StringVar(&ConfigFile, "c", "config.txt", "Specify a configuration file.")
}

func main() {
	flag.Parse()
	util.SetErrorConsoleLogging(ConsoleErrors)
	util.SetInfoConsoleLogging(ConsoleInfo)
	tool.SetTimeout(Timeout)
	err := config.LoadConfigs(ConfigFile)
	if err != nil {
		util.Log(err)
		return
	}
	if UsersFile != "" {
		AddUsers()
	}
	if Web {
		RunWebServer(Receiver || Processor)
	}
	if Processor {
		RunFileProcessor(Receiver)
	}
	if Receiver {
		RunFileReceiver(false)
	}
}

//AddUsers adds users from a text file to the database.
func AddUsers() {
	users, err := user.ReadUsers(UsersFile)
	if err != nil {
		util.Log(err)
		return
	}
	err = db.Setup(db.DEFAULT_CONN)
	if err != nil {
		util.Log(err)
		return
	}
	err = db.AddUsers(users...)
	if err != nil {
		util.Log(err)
	}

}

func RunWebServer(inRoutine bool) {
	err := db.Setup(db.DEFAULT_CONN)
	if err != nil {
		util.Log(err)
	} else {
		if inRoutine {
			go webserver.Run()
		} else {
			webserver.Run()
		}
	}
}

func RunFileReceiver(inRoutine bool) {
	err := db.Setup(db.DEFAULT_CONN)
	if err != nil {
		util.Log(err)
	} else {
		if inRoutine {
			go server.Run(Port, new(server.SubmissionSpawner))
		} else {
			server.Run(Port, new(server.SubmissionSpawner))
		}
	}
}

func RunFileProcessor(inRoutine bool) {
	err := db.Setup(db.DEFAULT_CONN)
	if err != nil {
		util.Log(err)
	} else {
		if inRoutine {
			go processing.Serve(MaxProcs)
		} else {
			processing.Serve(MaxProcs)
		}
	}
}
