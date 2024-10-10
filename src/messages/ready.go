package messages

import (
	"fmt"
	"net"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type ReadyMessage struct {
	Set struct {
		Ready struct {
			Username          string `json:"username"`
			IsReady           bool   `json:"isReady"`
			ManuallyInitiated bool   `json:"manuallyInitiated"`
		} `json:"ready"`
	} `json:"Set"`
}

type ClientReadyMessage struct {
	IsReady           bool `json:"isReady"`
	ManuallyInitiated bool `json:"manuallyInitiated"`
}

func SendReadyMessageInit(conn net.Conn, username string) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	if room == nil {
		return
	}

	room.SetUserReadyState(username, false, false)

	readyMessage := ReadyMessage{
		Set: struct {
			Ready struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			} `json:"ready"`
		}{
			Ready: struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			}{
				Username:          username,
				IsReady:           false,
				ManuallyInitiated: false,
			},
		},
	}

	utils.SendJSONMessageMultiCast(readyMessage, room.Name)
}

func HandleReadyMessage(ready map[string]interface{}, conn net.Conn) {
	// Print the incoming message
	cm := connM.GetConnectionManager()

	// Unmarshal the incoming JSON data into ClientReadyMessage struct
	var clientReadyMessage ClientReadyMessage

	// Extract the isReady and manuallyInitiated
	isReady := clientReadyMessage.IsReady
	manuallyInitiated := clientReadyMessage.ManuallyInitiated

	// Get the room and user associated with the connection
	room := cm.GetRoomByConnection(conn)
	if room == nil {
		fmt.Println("Error: Room not found for connection")
		return
	}

	// Assuming username is extracted from the connection or another source
	username := room.GetUsernameByConnection(conn)
	if username == "" {
		fmt.Println("Error: Username not found for connection")
		return
	}

	// Set the user ready state
	room.SetUserReadyState(username, isReady, manuallyInitiated)

	// Send the ready message to all connections in the room
	readyMessage := map[string]interface{}{
		"Set": map[string]interface{}{
			"ready": map[string]interface{}{
				"username":          username,
				"isReady":           isReady,
				"manuallyInitiated": manuallyInitiated,
			},
		},
	}

	room.PrintReadyStates()

	// Send the ready message to all connections in the room

	utils.SendJSONMessageMultiCast(readyMessage, room.Name)
}
