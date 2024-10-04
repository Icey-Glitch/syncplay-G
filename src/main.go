package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/Icey-Glitch/Syncplay-G/ConMngr"
	"github.com/Icey-Glitch/Syncplay-G/messages"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

func main() {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			fmt.Println("Error closing listener:", err)
		}
	}(ln)

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
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			cm := ConMngr.GetConnectionManager()
			fmt.Println("Client disconnected")
			cm.RemoveConnection(conn)
			return
		} else if err != nil {
			fmt.Println("Error decoding message:", err)
			return
		}

		if msg == nil {
			fmt.Println("Empty message")
			continue
		}

		switch {
		case msg["TLS"] != nil:
			handleStartTLSMessage(conn)
		case msg["Hello"] != nil:
			handleHelloMessage(msg["Hello"], encoder, conn)
		case msg["State"] != nil:
			handleStateMessage(msg["State"], encoder, conn)
		case msg["Chat"] != nil:
			handleChatMessage(msg["Chat"], encoder, conn)
		case msg["Set"] != nil:
			messages.HandleReadyMessage(msg["Set"].(map[string]interface{}), conn)
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

	//version, ok := helloData["version"].(string)
	if !ok {
		fmt.Println("Error: version is not a string")
		return
	}

	room := helloData["room"].(map[string]interface{})

	cm := ConMngr.GetConnectionManager()
	cm.AddConnection(username, nil, conn)

	sendSessionInformation(conn, username)

	response := messages.CreateHelloResponse(username, "1.7.3", room["name"].(string))

	utils.SendJSONMessage(conn, response)

	messages.SendInitialState(conn, username)

	//utils.PrettyPrintJSON(jsonData)
}

func handleStateMessage(stateMsg interface{}, encoder *json.Encoder, conn net.Conn) {
	var position, paused, doSeek, setBy interface{}
	var messageAge, latencyCalculation float64
	clientIgnoringOnTheFly := 0

	stateData, ok := stateMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: State message is not a map")
		return
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
		position, paused, doSeek, setBy = messages.ExtractStatePlaystateArguments(playstate)
	}

	if ping, ok := stateData["ping"].(map[string]interface{}); ok {
		messageAge, latencyCalculation = messages.HandleStatePing(ping)
	}

	if position != nil && paused != nil && clientIgnoringOnTheFly == 0 {
		messages.UpdateGlobalState(position, paused, doSeek, setBy, messageAge)
	}

	position, paused, doSeek, stateChange := messages.GetLocalState()
	messages.SendStateMessage(conn, position, paused, doSeek, latencyCalculation, stateChange)
}

func handleChatMessage(chatData interface{}, encoder *json.Encoder, conn net.Conn) {
	cm := ConMngr.GetConnectionManager()
	// chatData is expected to be a map with a "Chat" key
	msg, ok := chatData.(string)
	if !ok {
		fmt.Println("Error decoding chat message: chatData is not a map")
		fmt.Printf("chatData type: %T\n", chatData)
		return
	}

	// Assuming chatStr is the actual chat message
	username := cm.GetUsername(conn)

	messages.SendChatMessage(msg, username)
}

func sendSessionInformation(conn net.Conn, username string) {
	messages.SendReadyMessageInit(conn, username)
	messages.SendPlaylistChangeMessage(conn)
	messages.SendPlaylistIndexMessage(conn)
}
