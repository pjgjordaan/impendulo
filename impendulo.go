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
	"github.com/godfried/impendulo/processor"
	"github.com/godfried/impendulo/processor/monitor"
	"github.com/godfried/impendulo/processor/mq"
	"github.com/godfried/impendulo/receiver"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/web"
	"labix.org/v2/mgo/bson"

	"os"
	"runtime"
	"strconv"
	"strings"
)

//Flag variables for setting ports to listen on, users file to process, mode to run in, etc.
var (
	wFlags, rFlags, pFlags   *flag.FlagSet
	cfgFile, errLog, infoLog string
	backupDB, access         string
	dbName, dbAddr, mqURI    string
	mProcs                   int
	httpPort, tcpPort        int
)

func init() {
	d, e := config.DefaultConfig()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	//Setup flags
	flag.StringVar(&backupDB, "b", "", "Specify a backup db (default none).")
	flag.StringVar(&errLog, "e", "a", "Specify where to log errors to (default console & file).")
	flag.StringVar(&infoLog, "i", "f", "Specify where to log info to (default file).")
	flag.StringVar(&cfgFile, "c", d, fmt.Sprintf("Specify a configuration file (default %s).", d))
	flag.StringVar(&dbName, "db", db.DEBUG_DB, fmt.Sprintf("Specify a db to use (default %s).", db.DEBUG_DB))
	flag.StringVar(&dbAddr, "da", db.ADDRESS, fmt.Sprintf("Specify a db address to use (default %s).", db.ADDRESS))
	flag.StringVar(&access, "a", "",
		"Change a user's access permissions."+
			"Available permissions: NONE=0, STUDENT=1, TEACHER=2, ADMIN=3."+
			"Example: -a=pieter:2.")
	flag.StringVar(&mqURI, "mq", mq.DEFAULT_AMQP_URI, fmt.Sprintf("Specify the address of the Rabbitmq server (default %s).", mq.DEFAULT_AMQP_URI))

	pFlags = flag.NewFlagSet("processor", flag.ExitOnError)
	rFlags = flag.NewFlagSet("receiver", flag.ExitOnError)
	wFlags = flag.NewFlagSet("web", flag.ExitOnError)

	pFlags.IntVar(&mProcs, "mp", processor.MAX_PROCS, fmt.Sprintf("Specify the maximum number of goroutines to run when processing submissions (default %d).", processor.MAX_PROCS))

	rFlags.IntVar(&tcpPort, "p", receiver.PORT, fmt.Sprintf("Specify the port to listen on for files using TCP (default %d).", receiver.PORT))
	wFlags.IntVar(&httpPort, "p", web.PORT, fmt.Sprintf("Specify the port to use for the webserver (default %d).", web.PORT))
}

func main() {
	var e error
	defer func() {
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			util.Log(e)
		}
	}()
	//Set the number of processors to use to the number of available CPUs.
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	util.SetErrorLogging(errLog)
	util.SetInfoLogging(infoLog)
	mq.SetAMQP_URI(mqURI)
	//Handle setup flags
	if e = backup(backupDB); e != nil {
		return
	}
	if e = setupConn(dbAddr, dbName); e != nil {
		return
	}
	defer db.Close()
	if e = modifyAccess(access); e != nil {
		return
	}
	if flag.NArg() < 1 {
		e = fmt.Errorf("too few arguments provided %d", flag.NArg())
		return
	}
	if e = config.LoadConfigs(cfgFile); e != nil {
		return
	}
	switch flag.Arg(0) {
	case "web":
		runWebServer(httpPort)
	case "receiver":
		runFileReceiver(tcpPort)
	case "processor":
		runFileProcessor(mProcs)
	}
}

//modifyAccess changes a specified user's access permissions.
//Modification is specified as username:new_permission_level where
//new_permission_level can be integers from 0 to 3.
func modifyAccess(a string) error {
	if a == "" {
		return nil
	}
	ps := strings.Split(a, ":")
	if len(ps) != 2 {
		return fmt.Errorf("invalid parameters %s for user access modification", a)
	}
	v, e := strconv.Atoi(ps[1])
	if e != nil {
		return fmt.Errorf("invalid user access token %s", ps[1])
	}
	p := user.Permission(v)
	if p < user.NONE || p > user.ADMIN {
		return fmt.Errorf("invalid user access token %d", v)
	}
	if e = db.Update(db.USERS, bson.M{user.ID: ps[0]}, bson.M{db.SET: bson.M{user.ACCESS: p}}); e != nil {
		return fmt.Errorf("update error: user %s's access permissions", ps[0])
	}
	fmt.Printf("updated %s's permission level to %s\n", ps[0], p.Name())
	return nil
}

//backup backs up the default database to a specified backup.
func backup(b string) error {
	if b == "" {
		return nil
	}
	if e := db.Setup(db.DEFAULT_CONN); e != nil {
		return e
	}
	defer db.Close()
	if e := db.CopyDB(db.DEFAULT_DB, b); e != nil {
		return e
	}
	fmt.Printf("successfully backed-up main db to %s.\n", b)
	return nil
}

//setupConn sets up the database connection
func setupConn(a, n string) error {
	return db.Setup(a + n)
}

//runWebServer runs the webserver
func runWebServer(p int) {
	wFlags.Parse(os.Args[2:])
	web.Run(p)
}

//runFileReceiver runs the TCP file receiving server.
func runFileReceiver(p int) {
	rFlags.Parse(os.Args[2:])
	receiver.Run(tcpPort, new(receiver.SubmissionSpawner))
}

//runFileProcessor runs the file processing server.
func runFileProcessor(n int) {
	pFlags.Parse(os.Args[2:])
	go monitor.Start()
	processor.Serve(n)
}
