package main

import (
	"flag"
	"github.com/disco-volante/intlola/client"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/server"
	"github.com/disco-volante/intlola/utils"
)

var port, address, users string

func init() {
	flag.StringVar(&port, "p", "9999", "Specify the port to listen on.")
	flag.StringVar(&address, "a", "0.0.0.0", "Specify the address.")
	flag.StringVar(&users, "u", "", "Specify a file with new users.")

}

func main() {
	flag.Parse()
	initDB()
	if users != ""{
		addUsers(users)
	}
	runServer(address, port)
}

func runServer(addr, port string) {
	utils.Log("Starting server at: ", address, " on port ", port)
	server.Run(addr, port)
}

func addUsers(fname string) {
	users, err := utils.ReadUsers(fname)
	if err != nil {
		utils.Log("Invalid users file ", fname, " gave error: ", err)
	} else {
		for user, pword := range users {
			data := &client.ClientData{make(map[string]int), pword}
			db.Add(user, data)
		}
	}
}

func initDB() {
	utils.MkDir("db")
}
