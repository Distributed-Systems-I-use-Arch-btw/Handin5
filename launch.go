package main

import (
	client "Handin5/client"
	server "Handin5/server"
	"fmt"
	"os"
)

func main() {
	arg := os.Args

	if len(arg) < 2 {
		fmt.Println("Please specify 'client' or 'server' as the only argument.")
		os.Exit(1)
	}

	switch arg[1] {
	case "client":
		client.Run()
	case "server":
		server.Run()
	default:
		fmt.Printf("Unknown argument: %s. Use 'client' or 'server'.\n", arg[1])
		os.Exit(1)
	}
}
