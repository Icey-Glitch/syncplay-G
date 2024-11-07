package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	Features "github.com/Icey-Glitch/Syncplay-G/features"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"

	"github.com/goccy/go-json"

	"github.com/Icey-Glitch/Syncplay-G/messages"
	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

var (
	maxWorkers = 100 // limit concurrent goroutines
	workerPool = make(chan struct{}, maxWorkers)
)

func main() {
	features := Features.NewFeatures()
	Features.SetGlobalFeatures(*features)
	config := Features.NewConfig()
	Features.SetConfig(*config)

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
	err := conn.SetDeadline(time.Now().Add(time.Minute * 5))
	if err != nil {
		fmt.Println("Failed to set deadline" + err.Error())
		return
	}

	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)

	for {
		var msg map[string]interface{}
		if err = decoder.Decode(&msg); err == io.EOF {
			cm := connM.GetConnectionManager()
			fmt.Println("Client disconnected")
			room := cm.GetRoomByConnection(conn)
			if room == nil {
				return
			}

			usr, err := room.GetConnectionByConn(conn)
			if err != nil {
				//fmt.Println("Error getting connection by conn:", err)
				return
			}
			messages.HandleUserLeftMessage(*usr)
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
			handleHelloMessage(msg["Hello"], conn)
		case msg["State"] != nil:
			handleStateMessage(msg["State"], conn)
		case msg["Chat"] != nil:
			handleChatMessage(msg["Chat"], conn)
		case msg["Set"] != nil:
			handleSetMessage(msg["Set"], conn)
		case msg["List"] == nil:
			handleListMessage(conn)
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

	/*
		jsonData := []byte(`{"TLS": {"startTLS": "false"}}`)
	*/

	// erase user with duplicate connection if the connection makes another startTLS request
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	if room != nil {
		fmt.Println("Removing connection from room")
		cm.RemoveConnection(conn)
	}

	fmt.Println("Sending StartTLS response:", string(payload))
	if _, err := conn.Write(payload); err != nil {
		fmt.Println("Error sending StartTLS response:", err)
	}

}

func handleHelloMessage(helloMsg interface{}, conn net.Conn) {
	helloData, ok := helloMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: room is not a map, got:", helloData["room"])
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
		roomObj := cm.CreateRoom(roomName)
		if roomObj == nil {
			fmt.Println("Failed to create room")
			return
		}
	}

	connection, coner := cm.AddConnection(username, roomName, nil, conn)
	if coner != nil {
		fmt.Println("Error adding connection to room:", coner)
		messages.SendMessageToUser(username+" Is already in the room", "server", conn)
		return
	}

	err1 := messages.BroadcastJoinAnnouncement(*connection)
	if err1 != nil {
		fmt.Println("Failed to send Join Anouncement" + err1.Error())
		return
	}

	sendSessionInformation(*connection)

	response := messages.CreateHelloResponse(username, "1.7.3", roomName)
	err := utils.SendJSONMessage(conn, response)
	if err != nil {
		fmt.Println("failed to send hello to " + username + " " + err.Error())
		return
	}

	messages.SendInitialState(*connection)

	setupStatusScheduler(*connection)

}

func handleSetMessage(setMsg interface{}, conn net.Conn) {
	// Deserialize the set message
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	usr, err := room.GetConnectionByConn(conn)
	if err != nil {
		//fmt.Println("Error getting connection by conn:", err)
		return
	}

	setData, ok := setMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: Set message is not a map")
		return
	}

	fmt.Println("Set message:", setData)

	for key, value := range setData {
		switch key {
		case "user":
			messages.HandleUserMessage(value, conn)
		case "ready":
			messages.HandleReadyMessage(value, usr)
		case "playlistChange":
			messages.HandlePlaylistChangeMessage(value, *usr)
		case "playlistIndex":
			messages.HandlePlaylistIndexMessage(*usr, value)
		case "file":
			messages.HandleFileMessage(*usr, value)
		case "room":
			messages.HandleUserMoveRoomMessage(*usr, value)
		default:
			fmt.Printf("Unknown message type: %s\n", key)
		}
	}
}

// func handle list message
func handleListMessage(conn net.Conn) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	usr, err := room.GetConnectionByConn(conn)
	if err != nil {
		//fmt.Println("Error getting connection by conn:", err)
		return
	}

	messages.HandleListRequest(*usr)

}

func handleStateMessage(stateMsg interface{}, conn net.Conn) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	user, err := room.GetConnectionByConn(conn)
	if err != nil {
		//fmt.Println("Error getting connection by conn:", err)
		return
	}
	var position, paused, doSeek, setBy interface{}
	var messageAge, latencyCalculation, clientlatency, rtt float64
	var clientIgnoringOnTheFly float64

	stateData, ok := stateMsg.(map[string]interface{})
	if !ok {
		fmt.Println("Error: State message is not a map")
		return
	}

	//Client >> {"State": {"ignoringOnTheFly": {"client": 1}, "ping": {"clientRtt": 0, "clientLatencyCalculation": 1394587857.902}, "playstate": {"paused": false, "position": 0.089}}}
	if ignoringOnTheFly, ok := stateData["ignoringOnTheFly"].(map[string]interface{}); ok {
		clientIgnoringOnTheFly, ok = ignoringOnTheFly["client"].(float64)
		if !ok {
			clientIgnoringOnTheFly = 0
		}

		fmt.Println("Ignoring on the fly:", clientIgnoringOnTheFly)
	}

	if playstate, ok := stateData["playstate"].(map[string]interface{}); ok {
		position, paused, doSeek, setBy = messages.ExtractStatePlaystateArguments(playstate, *user)
	}

	if ping, ok := stateData["ping"].(map[string]interface{}); ok {
		rtt, latencyCalculation, clientlatency = messages.HandleStatePing(ping)
		// store latency calculation in room
		err := room.SetUserLatencyCalculation(user, float64(time.Now().UnixNano())/1e9, clientlatency, rtt, latencyCalculation)
		if err != nil {
			fmt.Println("Error storing user latency calculation")
		}

	}

	if position != nil && paused != nil && doSeek != nil && setBy != nil {
		messages.UpdateGlobalState(*user, position, paused, doSeek, setBy, latencyCalculation, messageAge, clientIgnoringOnTheFly)
	}

	// if ignoringOnTheFly is not 0, send global state to all users
	if clientIgnoringOnTheFly != 0 {
		messages.SendGlobalState(*user)
	}

	messages.GetLocalState()

}

func handleChatMessage(chatData interface{}, conn net.Conn) {
	fmt.Println("Handling chat message")
	cm := connM.GetConnectionManager()
	msg, ok := chatData.(string)
	if !ok {
		fmt.Println("Error decoding chat message: chatData is not a string")
		return
	}

	username := cm.GetRoomByConnection(conn).GetUsernameByConnection(conn)
	messages.SendChatMessage(msg, username)
}

func sendSessionInformation(connection roomM.Connection) {
	messages.SendReadyMessageInit(connection)
	messages.SendPlaylistChangeMessage(connection, nil)
	messages.SendPlaylistIndexMessage(connection)
}

func setupStatusScheduler(connection roomM.Connection) {
	fmt.Println("Setting up status scheduler")
	room := connection.Owner
	if room == nil {
		fmt.Println("Error: Room not found")
		return
	}

	em := room.GetStateEventManager()

	params := []interface{}{connection}

	managedEvent := em.NewManagedEvent(1, messages.SendUserState, true, params, room.GetStateEventTicker())

	managedEvent.Start()
	fmt.Println("Status scheduler started")
}
