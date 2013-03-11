package main

import (
	"intlola/server"
	"intlola/db"
	"intlola/utils"
"intlola/client"
)

func main() {
	runServer("localhost", "9998")
	//addUsers("users")
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
			data := &client.ClientData{make(map[string] bool), pword}
			db.Add(user, data)
		}
	}
}
