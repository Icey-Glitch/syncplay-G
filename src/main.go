package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
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
		Username string `json:"username"`
		Room     struct {
			Name string `json:"name"`
		} `json:"room"`
		Version     string   `json:"version"`
		RealVersion string   `json:"realversion"`
		Features    Features `json:"features"`
		MOTD        string   `json:"motd"`
	} `json:"Hello"`
}

type Features struct {
	IsolateRooms         bool `json:"isolateRooms"`
	Readiness            bool `json:"readiness"`
	ManagedRooms         bool `json:"managedRooms"`
	PersistentRooms      bool `json:"persistentRooms"`
	Chat                 bool `json:"chat"`
	MaxChatMessageLength int  `json:"maxChatMessageLength"`
	MaxUsernameLength    int  `json:"maxUsernameLength"`
	MaxRoomNameLength    int  `json:"maxRoomNameLength"`
	MaxFilenameLength    int  `json:"maxFilenameLength"`
}

// Define structures for session set messages
type ReadyMessage struct {
	Set struct {
		Ready struct {
			Username          string      `json:"username"`
			IsReady           interface{} `json:"isReady"`
			ManuallyInitiated bool        `json:"manuallyInitiated"`
		} `json:"ready"`
	} `json:"Set"`
}

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			Files []interface{} `json:"files"`
			User  interface{}   `json:"user"`
		} `json:"playlistChange"`
	} `json:"Set"`
}

type PlaylistIndexMessage struct {
	Set struct {
		PlaylistIndex struct {
			Index interface{} `json:"index"`
			User  interface{} `json:"user"`
		} `json:"playlistIndex"`
	} `json:"Set"`
}

// Define the StateMessage structure
type StateMessage struct {
	State struct {
		Ping struct {
			LatencyCalculation float64 `json:"latencyCalculation"`
			ServerRtt          int     `json:"serverRtt"`
		} `json:"ping"`
		Playstate struct {
			Position float64     `json:"position"`
			Paused   bool        `json:"paused"`
			DoSeek   bool        `json:"doSeek"`
			SetBy    interface{} `json:"setBy"`
		} `json:"playstate"`
	} `json:"State"`
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

		fmt.Printf("Received message: %v\n", msg)

		switch {
		case msg["TLS"] != nil:
			handleStartTLSMessage(conn)
		case msg["Hello"] != nil:
			handleHelloMessage(msg["Hello"], encoder, conn)
		case msg["State"] != nil:
			handleStateMessage(msg["State"], encoder, conn)
		default:
			fmt.Println("Unknown message type")
			fmt.Println("Message:", msg)
		}
	}
}

func handleStartTLSMessage(conn net.Conn) {
	payload := []byte{
		0x7b, 0x22, 0x54, 0x4c, 0x53, 0x22, 0x3a, 0x20, 0x7b, 0x22, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54,
		0x4c, 0x53, 0x22, 0x3a, 0x20, 0x22, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x22, 0x7d, 0x7d, 0x0d, 0x0a,
	}

	fmt.Printf("Sending StartTLS response: %x\n", payload)
	if _, err := conn.Write(payload); err != nil {
		fmt.Println("Error sending StartTLS response:", err)
	}
}

func handleHelloMessage(helloMsg interface{}, encoder *json.Encoder, conn net.Conn) {
	helloData, ok := helloMsg.(map[string]interface{})
	if !ok {
		fmt.Printf("Error: room is not a map, got: %T\n", helloData["room"])
		return
	}

	username, ok := helloData["username"].(string)
	if !ok {
		fmt.Println("Error: username is not a string")
		return
	}

	version, ok := helloData["version"].(string)
	if !ok {
		fmt.Println("Error: version is not a string")
		return
	}

	room := helloData["room"].(map[string]interface{})

	sendSessionInformation(conn, username)

	response := createHelloResponse(username, version, room["name"].(string))

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling Hello response:", err)
		return
	}

	jsonData = insertSpaceAfterColons(jsonData)
	jsonData = append(jsonData, '\n')

	fmt.Printf("Sending Hello response: %s\n", jsonData)
	if _, err := conn.Write(jsonData); err != nil {
		fmt.Println("Error sending Hello response:", err)
	}

	sendInitialState(conn, username)

	prettyPrintJSON(jsonData)
}

func handleStateMessage(stateMsg interface{}, encoder *json.Encoder, conn net.Conn) {
	var position, paused, doSeek, setBy interface{}
	var messageAge, latencyCalculation float64
	hadFirstStateUpdate := false
	clientIgnoringOnTheFly := 0

	stateData, ok := stateMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: State message is not a map")
		return
	}

	if !hadFirstStateUpdate {
		hadFirstStateUpdate = true
	}

	if ignore, ok := stateData["ignoringOnTheFly"].(map[string]interface{}); ok {
		if _, ok := ignore["server"].(int); ok {
			clientIgnoringOnTheFly = 0
		} else if client, ok := ignore["client"].(int); ok {
			if client == clientIgnoringOnTheFly {
				clientIgnoringOnTheFly = 0
			}
		}
	}

	if playstate, ok := stateData["playstate"].(map[string]interface{}); ok {
		position, paused, doSeek, setBy = extractStatePlaystateArguments(playstate)
	}

	if ping, ok := stateData["ping"].(map[string]interface{}); ok {
		messageAge, latencyCalculation = handleStatePing(ping)
	}

	if position != nil && paused != nil && clientIgnoringOnTheFly == 0 {
		updateGlobalState(position, paused, doSeek, setBy, messageAge)
	}

	position, paused, doSeek, stateChange := getLocalState()
	sendState(position, paused, doSeek, latencyCalculation, stateChange, encoder)
}

func extractStatePlaystateArguments(playstate map[string]interface{}) (interface{}, interface{}, interface{}, interface{}) {
	position := playstate["position"]
	paused := playstate["paused"]
	doSeek := playstate["doSeek"]
	setBy := playstate["setBy"]
	return position, paused, doSeek, setBy
}

func handleStatePing(ping map[string]interface{}) (float64, float64) {
	messageAge := ping["messageAge"].(float64)
	latencyCalculation := ping["latencyCalculation"].(float64)
	return messageAge, latencyCalculation
}

func updateGlobalState(position, paused, doSeek, setBy interface{}, messageAge float64) {
	// Implement the logic to update the global state
}

func getLocalState() (interface{}, interface{}, interface{}, interface{}) {
	// Implement the logic to get the local state
	return nil, nil, nil, nil
}

func sendState(position, paused, doSeek, latencyCalculation, stateChange interface{}, encoder *json.Encoder) {
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = latencyCalculation.(float64)
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = doSeek.(bool)
	stateMessage.State.Playstate.Position = position.(float64)
	stateMessage.State.Playstate.Paused = paused.(bool)
	stateMessage.State.Playstate.SetBy = stateChange

	jsonData, err := json.Marshal(stateMessage)
	if err != nil {
		fmt.Println("Error marshaling state message:", err)
		return
	}

	jsonData = insertSpaceAfterColons(jsonData)
	jsonData = append(jsonData, '\n')

	if err := encoder.Encode(jsonData); err != nil {
		fmt.Println("Error sending state:", err)
	} else {
		fmt.Printf("Sent state message: %s\n", jsonData)
	}
}

func sendSessionInformation(conn net.Conn, username string) {
	sendReadyMessage(conn, username)
	sendPlaylistIndexMessage(conn)
	sendPlaylistChangeMessage(conn)
}

func sendReadyMessage(conn net.Conn, username string) {
	readyMessage := ReadyMessage{
		Set: struct {
			Ready struct {
				Username          string      `json:"username"`
				IsReady           interface{} `json:"isReady"`
				ManuallyInitiated bool        `json:"manuallyInitiated"`
			} `json:"ready"`
		}{
			Ready: struct {
				Username          string      `json:"username"`
				IsReady           interface{} `json:"isReady"`
				ManuallyInitiated bool        `json:"manuallyInitiated"`
			}{
				Username:          username,
				IsReady:           nil,
				ManuallyInitiated: false,
			},
		},
	}

	sendJSONMessage(conn, readyMessage)
}

func sendPlaylistIndexMessage(conn net.Conn) {
	playlistIndexMessage := PlaylistIndexMessage{
		Set: struct {
			PlaylistIndex struct {
				Index interface{} `json:"index"`
				User  interface{} `json:"user"`
			} `json:"playlistIndex"`
		}{
			PlaylistIndex: struct {
				Index interface{} `json:"index"`
				User  interface{} `json:"user"`
			}{
				Index: nil,
				User:  nil,
			},
		},
	}

	sendJSONMessage(conn, playlistIndexMessage)
}

func sendPlaylistChangeMessage(conn net.Conn) {
	playlistChangeMessage := PlaylistChangeMessage{
		Set: struct {
			PlaylistChange struct {
				Files []interface{} `json:"files"`
				User  interface{}   `json:"user"`
			} `json:"playlistChange"`
		}{
			PlaylistChange: struct {
				Files []interface{} `json:"files"`
				User  interface{}   `json:"user"`
			}{
				Files: []interface{}{},
				User:  nil,
			},
		},
	}

	sendJSONMessage(conn, playlistChangeMessage)
}

func sendInitialState(conn net.Conn, username string) {
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9 // Convert to seconds
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = 0
	stateMessage.State.Playstate.Paused = true
	stateMessage.State.Playstate.SetBy = nil // Initial state, no user set it

	sendJSONMessage(conn, stateMessage)
}

func sendJSONMessage(conn net.Conn, message interface{}) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}

	jsonData = insertSpaceAfterColons(jsonData)
	jsonData = append(jsonData, '\n')

	if _, err := conn.Write(jsonData); err != nil {
		fmt.Println("Error sending message:", err)
	} else {
		fmt.Printf("Sent message: %s\n", jsonData)
	}

	prettyPrintJSON(jsonData)
}

func insertSpaceAfterColons(jsonData []byte) []byte {
	modifiedData := make([]byte, 0, len(jsonData))
	for i := 0; i < len(jsonData); i++ {
		modifiedData = append(modifiedData, jsonData[i])
		if jsonData[i] == ':' {
			modifiedData = append(modifiedData, 0x20) // Insert space (0x20) after colon
		}
	}
	return modifiedData
}

func prettyPrintJSON(jsonData []byte) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, jsonData, "", "\t")
	if err != nil {
		fmt.Println("Error pretty printing JSON:", err)
		return
	}
	fmt.Println("Pretty printed JSON:", prettyJSON.String())
}

func createHelloResponse(username, version, roomName string) HelloResponseMessage {
	return HelloResponseMessage{
		Hello: struct {
			Username string `json:"username"`
			Room     struct {
				Name string `json:"name"`
			} `json:"room"`
			Version     string   `json:"version"`
			RealVersion string   `json:"realversion"`
			Features    Features `json:"features"`
			MOTD        string   `json:"motd"`
		}{
			Username: username,
			Room: struct {
				Name string `json:"name"`
			}{
				Name: roomName,
			},
			Version:     version,
			RealVersion: "1.7.3",
			Features: Features{
				IsolateRooms:         false,
				Readiness:            true,
				ManagedRooms:         true,
				PersistentRooms:      false,
				Chat:                 true,
				MaxChatMessageLength: 150,
				MaxUsernameLength:    16,
				MaxRoomNameLength:    35,
				MaxFilenameLength:    250,
			},
			MOTD: "",
		},
	}
}
