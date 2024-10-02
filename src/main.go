package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

var syncFactory SyncFactory

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		syncFactory.Port = "8080"
	} else {
		syncFactory.Port = os.Args[1]
		fmt.Println("Starting server on port", syncFactory.Port)
	}

	// set default values
	syncFactory.Password = ""
	syncFactory.Disable_chat = false
	syncFactory.Disable_ready = false

	listener()

}

// syncfactory comppatible arrgument struct
type SyncFactory struct {
	Port          string
	Password      string
	Disable_chat  bool
	Disable_ready bool
}

// Tcp listener
func listener() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

// Handle connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println("Message Received:", message)

		// twisted python compatible accept
		// use syncfactory struct to pass arguments like twisted
		/*
			client request:
			{"TLS": {"startTLS": "send"}}
		*/

		err = json.Unmarshal([]byte(message), &syncFactory)
		if err != nil {
			log.Println("Error decoding message:", err)
			continue
		}

		// Access the values from syncFactory struct
		port := syncFactory.Port
		password := syncFactory.Password
		disableChat := syncFactory.Disable_chat
		disableReady := syncFactory.Disable_ready

		// Use the values as needed
		fmt.Println("Port:", port)
		fmt.Println("Password:", password)
		fmt.Println("Disable Chat:", disableChat)
		fmt.Println("Disable Ready:", disableReady)

		// Send response
		conn.Write([]byte("Message Received\n"))

	}
}
