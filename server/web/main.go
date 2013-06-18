package main

import (
	"net/http"
	"github.com/godfried/impendulo/handlers"
)

func main() {
	if err := http.ListenAndServe(":"+"8080", handlers.Router); err != nil {
		panic(err)
	}
}
