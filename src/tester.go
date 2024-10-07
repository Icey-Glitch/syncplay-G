package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
)

// Decode client messages according to the Syncplay protocol
func decodeClientMessage(data []byte) (map[string]interface{}, error) {
	var decodedMessage map[string]interface{}
	err := json.Unmarshal(data, &decodedMessage)
	if err != nil {
		return nil, err
	}
	return decodedMessage, nil
}

func handleClientConnection(clientConn net.Conn, serverAddr string) {
	defer clientConn.Close()

	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer serverConn.Close()

	// Forward data from client to server
	go forwardData(clientConn, serverConn, true)

	// Forward data from server to client
	go forwardData(serverConn, clientConn, false)

	// Wait for either connection to close
	select {}
}

func forwardData(src net.Conn, dst net.Conn, isClient bool) {
	reader := bufio.NewReader(src)
	writer := bufio.NewWriter(dst)

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data:", err)
			}
			return
		}

		if isClient {
			decodedMessage, err := decodeClientMessage(data)
			if err != nil {
				fmt.Println("Failed to decode client message:", err)
			} else {
				fmt.Println("Client message:", decodedMessage)
			}
		} else {
			decodedMessage, err := decodeClientMessage(data)
			if err != nil {
				fmt.Println("Failed to decode server message:", err)
			} else {
				fmt.Println("Server message:", decodedMessage)
			}
		}

		_, err = writer.Write(data)
		if err != nil {
			fmt.Println("Failed to write data:", err)
			return
		}
		writer.Flush()
	}
}

func startProxyServer(listenAddr, serverAddr string) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Failed to start listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on", listenAddr)
	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handleClientConnection(clientConn, serverAddr)
	}
}
func main() {
	listenAddr := flag.String("listen", ":8080", "Address to listen on")
	serverAddr := flag.String("server", "localhost:12345", "Address of the real server")
	flag.Parse()

	startProxyServer(*listenAddr, *serverAddr)
}
