package main

import (
	"fmt"
	"github.com/Grog2903/pet-cache-service/service/server"
	"github.com/Grog2903/pet-cache-service/service/storage"
)

func main() {
	store := storage.NewStorage()
	serv := server.NewServer(store)

	err := serv.Start()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
