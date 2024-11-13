package main

import (
	"bufio"
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
	maxWorkers = 100                      // Limit concurrent workers
	connChan   = make(chan net.Conn, 100) // Channel to queue incoming connections
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
	defer ln.Close()

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			utils.DebugLog("Error accepting connection:", err)
			continue
		}

		// Send the connection to connChan
		connChan <- conn
	}
}

type HelloMessage struct {
	Username string   `json:"username"`
	Room     RoomInfo `json:"room"`
}

type RoomInfo struct {
	Name string `json:"name"`
}

type TLSMessage struct {
	StartTLS string `json:"startTLS"`
}

type Message struct {
	TLS   *TLSMessage                  `json:"TLS,omitempty"`
	Hello *HelloMessage                `json:"Hello,omitempty"`
	State *messages.ClientStateMessage `json:"State,omitempty"`
	Chat  string                       `json:"Chat,omitempty"`
	Set   *messages.SetMessage         `json:"Set,omitempty"`
	List  *messages.ListRequest        `json:"List,omitempty"`
}

func worker() {
	for conn := range connChan {
		handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	err := conn.SetDeadline(time.Now().Add(time.Minute * 5))
	if err != nil {
		utils.DebugLog("Failed to set deadline:", err)
		return
	}

	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)

	for {
		var msg Message
		if err = decoder.Decode(&msg); err == io.EOF {
			cm := connM.GetConnectionManager()
			utils.DebugLog("Client disconnected")
			room := cm.GetRoomByConnection(conn)
			if room == nil {
				return
			}

			usr, err := room.GetConnectionByConn(conn)
			if err != nil {
				return
			}
			messages.HandleUserLeftMessage(*usr)
			cm.RemoveConnection(conn)

			// Stop any associated ManagedEvent
			em := room.GetStateEventManager()
			em.StopAll()

			return
		} else if err != nil {
			utils.DebugLog("Error decoding message:", err)
			break
		}

		switch {
		case msg.TLS != nil:
			handleStartTLSMessage(conn)
		case msg.Hello != nil:
			handleHelloMessage(msg.Hello, conn)
		case msg.State != nil:
			handleStateMessage(msg.State, conn)
		case msg.Chat != "":
			handleChatMessage(msg.Chat, conn)
		case msg.Set != nil:
			handleSetMessage(msg.Set, conn)
		case msg.List == nil:
			handleListMessage(conn)
		default:
			utils.DebugLog("Unknown message type: %v\n", msg)
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
		utils.DebugLog("Removing connection from room")
		cm.RemoveConnection(conn)
	}

	utils.DebugLog("Sending StartTLS response:", string(payload))
	if _, err := conn.Write(payload); err != nil {
		utils.DebugLog("Error sending StartTLS response:", err)
	}

}

func handleHelloMessage(helloMsg *HelloMessage, conn net.Conn) {
	username := helloMsg.Username
	roomName := helloMsg.Room.Name

	cm := connM.GetConnectionManager()
	if cm.GetRoom(roomName) == nil {
		roomObj := cm.CreateRoom(roomName)
		if roomObj == nil {
			utils.DebugLog("Failed to create room")
			return
		}
	}

	connection, coner := cm.AddConnection(username, roomName, nil, conn)
	if coner != nil {
		utils.DebugLog("Error adding connection to room:", coner)
		messages.SendMessageToUser(username+" is already in the room", "server", conn)
		return
	}

	err := messages.BroadcastJoinAnnouncement(*connection)
	if err != nil {
		utils.DebugLog("Failed to send join announcement:", err)
		return
	}

	sendSessionInformation(*connection)

	response := messages.CreateHelloResponse(username, "1.7.3", roomName)
	err = utils.SendJSONMessage(conn, response)
	if err != nil {
		utils.DebugLog("Failed to send hello to", username, ":", err)
		return
	}

	messages.SendInitialState(*connection)

	setupStatusScheduler(*connection)
}

func handleSetMessage(setMsg *messages.SetMessage, conn net.Conn) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	usr, err := room.GetConnectionByConn(conn)
	if err != nil {
		return
	}

	switch {
	case setMsg.User != nil:
		messages.HandleUserMessage(setMsg.User, conn)
	case setMsg.Ready != nil:
		messages.HandleReadyMessage(setMsg.Ready, usr)
	case setMsg.PlaylistChange != nil:
		messages.HandlePlaylistChangeMessage(setMsg.PlaylistChange, *usr)
	case setMsg.PlaylistIndex != nil:
		messages.HandlePlaylistIndexMessage(*usr, setMsg.PlaylistIndex)
	case setMsg.File != nil:
		messages.HandleFileMessage(*usr, setMsg.File)
	case setMsg.Room != nil:
		messages.HandleUserMoveRoomMessage(*usr, setMsg.Room)
	}

}

// func handle list message
func handleListMessage(conn net.Conn) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	usr, err := room.GetConnectionByConn(conn)
	if err != nil {
		//utils.DebugLog("Error getting connection by conn:", err)
		return
	}

	messages.HandleListRequest(*usr)

}

func handleStateMessage(stateMsg *messages.ClientStateMessage, conn net.Conn) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	user, err := room.GetConnectionByConn(conn)
	if err != nil {
		return
	}

	// pritty print the state message
	utils.DebugLog("State message:", stateMsg)

	position := stateMsg.Playstate.Position
	paused := stateMsg.Playstate.Paused
	doSeek := stateMsg.Playstate.DoSeek
	setBy := stateMsg.Playstate.SetBy
	if setBy == "" || setBy == nil {
		setBy = "Nobody"
	}

	latencyCalculation := stateMsg.Ping.LatencyCalculation
	clientLatencyCalculation := stateMsg.Ping.ClientLatencyCalculation
	clientRtt := stateMsg.Ping.ClientRtt

	err = room.SetUserLatencyCalculation(user, float64(time.Now().UnixNano())/1e9, clientLatencyCalculation, clientRtt, latencyCalculation)
	if err != nil {
		utils.DebugLog("Error storing user latency calculation")
	}

	clientIgnoringOnTheFly := 0.0
	if stateMsg.IgnoringOnTheFly != nil {
		utils.DebugLog("Ignoring on the fly")
		clientIgnoringOnTheFly = stateMsg.IgnoringOnTheFly.Client
	}

	messages.UpdateGlobalState(*user, position, paused, doSeek, setBy, latencyCalculation, 0, clientIgnoringOnTheFly)

	if clientIgnoringOnTheFly != 0 {
		messages.SendGlobalState(*user)
	}
}

func handleChatMessage(chatMsg string, conn net.Conn) {
	utils.DebugLog("Handling chat message")
	cm := connM.GetConnectionManager()
	username := cm.GetRoomByConnection(conn).GetUsernameByConnection(conn)
	messages.SendChatMessage(chatMsg, username)
}

func sendSessionInformation(connection roomM.Connection) {
	messages.SendReadyMessageInit(connection)
	messages.SendPlaylistChangeMessage(connection, nil)
	messages.SendPlaylistIndexMessage(connection)
}

func setupStatusScheduler(connection roomM.Connection) {
	utils.DebugLog("Setting up status scheduler")
	room := connection.Owner
	if room == nil {
		utils.DebugLog("Error: Room not found")
		return
	}

	em := room.GetStateEventManager()

	params := []interface{}{connection}

	managedEvent := em.NewManagedEvent(1, messages.SendUserState, true, params, room.GetStateEventTicker())

	managedEvent.Start()
	utils.DebugLog("Status scheduler started")
}
