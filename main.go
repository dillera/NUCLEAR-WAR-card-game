package main

import (
	"fmt"
	"nuclear-war-game-server/api"
)

func main() {
	fmt.Println("Nuclear War Game Server Starting...")
	server := api.NewServer()
	server.Start()
}
