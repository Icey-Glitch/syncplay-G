package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/Icey-Glitch/Syncplay-G/messages"
	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

var (
	connPool = sync.Pool{
		New: func() interface{} {
			var conn net.Conn
			return &conn
		},
	}
	maxWorkers = 100 // limit concurrent goroutines
	workerPool = make(chan struct{}, maxWorkers)
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting server:", err)
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
	workerPool <- struct{}{} // limit goroutines
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
		<-workerPool // release worker
	}(conn)

	// Set a timeout to avoid stale connections
	conn.SetDeadline(time.Now().Add(time.Minute * 5))

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(writer)

	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err == io.EOF {
			cm := connM.GetConnectionManager()
			fmt.Println("Client disconnected")
			messages.HandleUserLeftMessage(conn)
			cm.RemoveConnection(conn)
			return
		} else if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
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
			handleSetMessage(msg["Set"], conn)
		case msg["List"] == nil:
			messages.HandleListRequest(conn, connM.GetConnectionManager().GetRoomByConnection(conn))
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

	fmt.Println("Sending StartTLS response: %x", payload)
	if _, err := conn.Write(payload); err != nil {
		fmt.Println("Error sending StartTLS response:", err)
	}
}

func handleHelloMessage(helloMsg interface{}, encoder *json.Encoder, conn net.Conn) {
	helloData, ok := helloMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: room is not a map, got: %T", helloData["room"])
		return
	}

	username, ok := helloData["username"].(string)
	if !ok {
		fmt.Println("Error: username is not a string")
		return
	}

	room, ok := helloData["room"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: room is not a map")
		return
	}

	roomName, ok := room["name"].(string)
	if !ok {
		fmt.Println("Error: room name is not a string")
		return
	}

	cm := connM.GetConnectionManager()
	if cm.GetRoom(roomName) == nil {
		cm.CreateRoom(roomName)
	}

	cm.AddConnection(username, roomName, nil, conn)

	sendSessionInformation(conn, username, roomName)

	response := messages.CreateHelloResponse(username, "1.7.3", roomName)
	utils.SendJSONMessage(conn, response)

	messages.SendInitialState(conn, username)

}

func handleSetMessage(setMsg interface{}, conn net.Conn) {
	// Deserialize the set message
	setData, ok := setMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: Set message is not a map")
		return
	}

	fmt.Println("Set message:", setData)
	// Handle user joining room
	if user, ok := setData["user"].(map[string]interface{}); ok {
		if user != nil {

			messages.HandleJoinMessage(conn, user)
		} else {
			fmt.Println("Error: user is nil")
		}
	}

	// Handle ready message
	if ready, ok := setData["ready"].(map[string]interface{}); ok {
		messages.HandleReadyMessage(ready, conn)
	}

	// Handle playlist change message
	if playlistChange, ok := setData["playlistChange"].(map[string]interface{}); ok {
		if playlistChange != nil {
			messages.HandlePlaylistChangeMessage(conn, playlistChange)
		} else {
			fmt.Println("Error: playlistChange is nil")
		}
	}

	// Handle playlist index message
	if playlistIndex, ok := setData["playlistIndex"].(map[string]interface{}); ok {
		if playlistIndex != nil {
			messages.HandlePlaylistIndexMessage(conn, playlistIndex)
		} else {
			fmt.Println("Error: playlistIndex is nil")
		}
	}

	// handle file message
	if file, ok := setData["file"].(map[string]interface{}); ok {
		if file != nil {
			messages.HandleFileMessage(conn, file)
		} else {
			fmt.Println("Error: file is nil")
		}
	}
}

func handleStateMessage(stateMsg interface{}, encoder *json.Encoder, conn net.Conn) {
	cm := connM.GetConnectionManager()
	user := cm.GetRoomByConnection(conn).GetUsernameByConnection(conn)
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
		position, paused, doSeek, setBy = messages.ExtractStatePlaystateArguments(playstate, conn)
	}

	if ping, ok := stateData["ping"].(map[string]interface{}); ok {
		messageAge, latencyCalculation = messages.HandleStatePing(ping)
		// store latency calculation in room
		cm.GetRoomByConnection(conn).SetUserLatencyCalculation(user, latencyCalculation)
	}

	if position != nil && paused != nil && clientIgnoringOnTheFly == 0 {
		messages.UpdateGlobalState(position, paused, doSeek, setBy, messageAge, latencyCalculation)
	}

	messages.GetLocalState()

}

func handleChatMessage(chatData interface{}, encoder *json.Encoder, conn net.Conn) {
	cm := connM.GetConnectionManager()
	msg, ok := chatData.(string)
	if !ok {
		fmt.Println("Error decoding chat message: chatData is not a string")
		return
	}

	username := cm.GetRoomByConnection(conn).GetUsernameByConnection(conn)
	messages.SendChatMessage(msg, username)
}

func sendSessionInformation(conn net.Conn, username, roomName string) {
	messages.SendReadyMessageInit(conn, username)
	messages.SendPlaylistChangeMessage(conn, roomName)
	messages.SendPlaylistIndexMessage(conn, roomName)
	go messages.AddUserToSchedule(conn, username)
}
