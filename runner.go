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
	initDB()
	addUsers(users)
	runServer(address, port)
}

func runServer(addr, port string) {
	server.Run(addr, port)
}

func addUsers(fname string) {
	utils.Log(fname)
	users, err := utils.ReadUsers(fname)
	if err != nil {
		utils.Log(err)
	} else {
		for user, pword := range users {
			data := &client.ClientData{make(map[string]int), pword}
			utils.Log(data)
			db.Add(user, data)
		}
	}
}

func initDB() {
	utils.MkDir("db")
}
