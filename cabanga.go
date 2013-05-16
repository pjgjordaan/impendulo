package main

import (
	"fmt"
	"flag"
	"github.com/godfried/cabanga/db"
	"github.com/godfried/cabanga/server"
	"github.com/godfried/cabanga/util"
	"log"
)

var port, users, mode string

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
		runServer(port)
	} else {
		log.Fatal(fmt.Errorf("Unknown running mode %q", mode))
	}
}

func AddUsers(fname string) error {
	users, err := util.ReadUsers(fname)
	if err != nil {
		return err
	}
	return db.AddUsers(users...)
}

func runServer(port string) {
	util.Log("Starting server on port ", port)
	server.Run(port)
}
