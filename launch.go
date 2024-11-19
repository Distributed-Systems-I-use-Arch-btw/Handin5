package main

import (
	client "Handin5/client"
	server "Handin5/server"
	"fmt"
	"net"
	"os"
	"time"
)

func start_server() {
	ports := []int{5050, 5051, 5052}

	for _, port := range ports {
		if checkPortAvailability(port) {
			server.Run(port)

			return
		}
	}

	fmt.Println("No available ports to start the client.")
}

func checkPortAvailability(port int) bool {
	_, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Second*2)
	if err != nil {
		return true
	}

	return false
}

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
		start_server()
	default:
		fmt.Printf("Unknown argument: %s. Use 'client' or 'server'.\n", arg[1])
		os.Exit(1)
	}
}
