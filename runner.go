package main

import (
	"flag"
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
	if users != ""{
		err := utils.AddUsers(users)
		if err != nil{
			utils.Log("DB error ", err)
		}
	}
	runServer(address, port)
}

func runServer(addr, port string) {
	utils.Log("Starting server at: ", address, " on port ", port)
	server.Run(addr, port)
}


