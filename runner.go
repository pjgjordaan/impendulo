package main

import (
	"github.com/disco-volante/intlola/server"
	"github.com/disco-volante/intlola/db"
	"github.com/disco-volante/intlola/utils"
"github.com/disco-volante/intlola/client"
)

func main() {
	if !checkDB(){
		addUsers("users")
	}
	runServer("localhost", "9998")
}

func runServer(addr, port string){
	server.Run(addr, port)
}

func addUsers(fname string){
	utils.MkDir("db")
	users, err := utils.ReadUsers(fname)
	if err != nil{
		utils.Log(err)
	} else{
		for user, pword := range users{
			data := &client.ClientData{make(map[string]int), pword}
			db.Add(user, data)
		}
	}
}

func checkDB() bool{
	return utils.MkDir("db") != nil
}