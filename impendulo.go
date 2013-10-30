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

//Package impendulo is provides storage and analysis for code snapshots.
//It receives code snapshots via TCP or a web upload, runs analysis tools and tests on
//them and provides a web interface to view the results in.
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
	"labix.org/v2/mgo/bson"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//Flag variables for setting ports to listen on, users file to process, mode to run in, etc.
var (
	wFlags, rFlags, pFlags               *flag.FlagSet
	cfgFile, errLog, infoLog             string
	backupDB, access                     string
	dbName, dbAddr, rpcAddr              string
	mProcs, timeout                      int
	httpPort, tcpPort, procPort, rpcPort int
)

const (
	LOG_IMPENDULO = "impendulo.go"
)

func init() {
	defualt, err := config.DefaultConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	//Setup flags
	flag.StringVar(&backupDB, "b", "", "Specify a backup db (default none).")
	flag.StringVar(&errLog, "e", "a", "Specify where to log errors to (default console & file).")
	flag.StringVar(&infoLog, "i", "f", "Specify where to log info to (default file).")
	flag.StringVar(&cfgFile, "c", defualt, "Specify a configuration file.")
	flag.StringVar(&dbName, "db", db.DEBUG_DB, "Specify a db to use (default "+db.DEBUG_DB+").")
	flag.StringVar(&dbAddr, "da", db.ADDRESS, "Specify a db address to use (default "+db.ADDRESS+").")
	flag.StringVar(
		&access, "a", "",
		"Change a user's access permissions."+
			"Available permissions: NONE=0, STUDENT=1, TEACHER=2, ADMIN=3."+
			"Example: -a=pieter:2.",
	)
	pFlags = flag.NewFlagSet("processor", flag.ExitOnError)
	rFlags = flag.NewFlagSet("receiver", flag.ExitOnError)
	wFlags = flag.NewFlagSet("web", flag.ExitOnError)

	pFlags.IntVar(&timeout, "t", 5, "Specify the time limit for a tool to run in, in minutes (default 5).")
	pFlags.IntVar(&mProcs, "mp", 5, "Specify the maximum number of goroutines to run when processing submissions (default 5).")
	pFlags.IntVar(&procPort, "p", 8045, "Specify the port to use for the processing server (default 8045).")

	rFlags.IntVar(&tcpPort, "p", 8010, "Specify the port to listen on for files using TCP (default 8010).")
	rFlags.StringVar(&rpcAddr, "ra", "", "Specify a server rpc address to connect to (default localhost).")
	rFlags.IntVar(&rpcPort, "rp", 8045, "Specify a server rpc port to connect to (default 8045).")

	wFlags.IntVar(&httpPort, "p", 8080, "Specify the port to use for the webserver (default 8080).")
	wFlags.StringVar(&rpcAddr, "ra", "", "Specify a server rpc address to connect to (default localhost).")
	wFlags.IntVar(&rpcPort, "rp", 8045, "Specify a server rpc port to connect to (default 8045).")

}

func main() {
	var err error
	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			util.Log(err)
		}
	}()
	//Set the number of processors to use to the number of available CPUs.
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	util.SetErrorLogging(errLog)
	util.SetInfoLogging(infoLog)
	//Handle setup flags
	if err = backup(); err != nil {
		return
	}
	if err = setupConn(); err != nil {
		return
	}
	defer db.Close()
	if err = modifyAccess(); err != nil {
		return
	}
	if flag.NArg() < 1 {
		return
	}
	if err = config.LoadConfigs(cfgFile); err != nil {
		return
	}
	switch flag.Arg(0) {
	case "web":
		runWebServer()
	case "receiver":
		runFileReceiver()
	case "processor":
		runFileProcessor()
	}
}

//modifyAccess changes a specified user's access permissions.
//Modification is specified as username:new_permission_level where
//new_permission_level can be integers from 0 to 3.
func modifyAccess() error {
	if access == "" {
		return nil
	}
	params := strings.Split(access, ":")
	if len(params) != 2 {
		return fmt.Errorf("Invalid parameters %s for user access modification.", access)
	}
	val, err := strconv.Atoi(params[1])
	if err != nil {
		return fmt.Errorf("Invalid user access token %s.", params[1])
	}
	newPerm := user.Permission(val)
	if newPerm < user.NONE || newPerm > user.ADMIN {
		return fmt.Errorf("Invalid user access token %d.", val)
	}

	change := bson.M{db.SET: bson.M{user.ACCESS: newPerm}}
	err = db.Update(db.USERS, bson.M{user.ID: params[0]}, change)
	if err != nil {
		return fmt.Errorf("Could not update user %s's access permissions.", params[0])
	}
	fmt.Printf("Updated %s's permission level to %s.\n", params[0], newPerm.Name())
	return nil
}

//backup backs up the default database to a specified backup.
func backup() (err error) {
	if backupDB == "" {
		return
	}
	err = db.Setup(db.DEFAULT_CONN)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.CopyDB(db.DEFAULT_DB, backupDB)
	if err == nil {
		fmt.Printf("Successfully Backed-up Main Database to %s.\n", backupDB)
	}
	return
}

//setupConn sets up the database connection
func setupConn() (err error) {
	err = db.Setup(dbAddr + dbName)
	return
}

//runWebServer runs the webserver
func runWebServer() {
	wFlags.Parse(os.Args[2:])
	processing.SetClientAddress(rpcAddr, rpcPort)
	webserver.Run(httpPort)
}

//runFileReceiver runs the TCP file receiving server.
func runFileReceiver() {
	rFlags.Parse(os.Args[2:])
	processing.SetClientAddress(rpcAddr, rpcPort)
	server.Run(tcpPort, new(server.SubmissionSpawner))
}

//runFileProcessor runs the file processing server.
func runFileProcessor() {
	pFlags.Parse(os.Args[2:])
	tool.SetTimeout(timeout)
	processing.Serve(procPort, mProcs)
}
