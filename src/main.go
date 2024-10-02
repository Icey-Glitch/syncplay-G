package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Define structures for different client message types
type HelloMessage struct {
	Hello struct {
		Features struct {
			Chat            bool   `json:"chat"`
			FeatureList     bool   `json:"featureList"`
			ManagedRooms    bool   `json:"managedRooms"`
			PersistentRooms bool   `json:"persistentRooms"`
			Readiness       bool   `json:"readiness"`
			SharedPlaylists bool   `json:"sharedPlaylists"`
			UIMode          string `json:"uiMode"`
		} `json:"features"`
		RealVersion string `json:"realversion"`
		Room        struct {
			Name string `json:"name"`
		} `json:"room"`
		Username string `json:"username"`
		Version  string `json:"version"`
	} `json:"Hello"`
}

// Define structures for different server message types
type HelloResponseMessage struct {
	Hello struct {
		Features struct {
			Chat                 bool `json:"chat"`
			IsolateRooms         bool `json:"isolateRooms"`
			ManagedRooms         bool `json:"managedRooms"`
			MaxChatMessageLength int  `json:"maxChatMessageLength"`
			MaxFilenameLength    int  `json:"maxFilenameLength"`
			MaxRoomNameLength    int  `json:"maxRoomNameLength"`
			MaxUsernameLength    int  `json:"maxUsernameLength"`
			PersistentRooms      bool `json:"persistentRooms"`
			Readiness            bool `json:"readiness"`
		} `json:"features"`
		MOTD        string `json:"motd"`
		RealVersion string `json:"realversion"`
		Room        struct {
			Name string `json:"name"`
		} `json:"room"`
		Username string `json:"username"`
		Version  string `json:"version"`
	} `json:"Hello"`
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			fmt.Println("Client disconnected")
			return
		} else if err != nil {
			fmt.Println("Error decoding message:", err)
			return
		}

		if msg == nil {
			fmt.Println("Empty message")
			continue
		}

		if msg["TLS"] != nil {
			handleStartTLSMessage(msg, encoder)
			continue
		}

		if msg["Hello"] != nil {
			handleHelloMessage(msg, encoder)
			continue
		}

	}
}

func handleStartTLSMessage(startTLSMsg interface{}, encoder *json.Encoder) {
	// Create the response message
	/*sample response
		Client message: map[TLS:map[startTLS:send]]
	Server message: map[TLS:map[startTLS:false]]
	*/
	response := map[string]interface{}{
		"TLS": map[string]interface{}{
			"startTLS": "false",
		},
	}

	fmt.Println("Server message:", response)
	fmt.Println("Client message:", startTLSMsg)

	//time.Sleep(10 * time.Second)
	encoder.Encode(response)

	// Send the response message

}

func handleHelloMessage(helloMsg interface{}, encoder *json.Encoder) {
	// Parse the incoming Hello message
	helloData := helloMsg.(map[string]interface{})
	_ = helloData["features"].(map[string]interface{})
	room := helloData["room"].(map[string]interface{})
	username := helloData["username"].(string)
	_ = helloData["version"].(string)

	// Create the response message
	response := HelloResponseMessage{}
	response.Hello.Features.Chat = true
	response.Hello.Features.IsolateRooms = false
	response.Hello.Features.ManagedRooms = true
	response.Hello.Features.MaxChatMessageLength = 150
	response.Hello.Features.MaxFilenameLength = 250
	response.Hello.Features.MaxRoomNameLength = 35
	response.Hello.Features.MaxUsernameLength = 16
	response.Hello.Features.PersistentRooms = false
	response.Hello.Features.Readiness = true
	response.Hello.RealVersion = "1.7.3"
	response.Hello.Room.Name = room["name"].(string)
	response.Hello.Username = username
	response.Hello.Version = "1.7.3"

	encoder.Encode(response)
}

func main() {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn)
	}
}
