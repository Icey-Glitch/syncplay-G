package messages

import (
	"fmt"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
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

func SendReadyMessageInit(connection roomM.Connection) {
	room := connection.Owner
	if room == nil {
		return
	}

	room.SetUserReadyState(connection.Username, false, false)

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
				Username:          connection.Username,
				IsReady:           false,
				ManuallyInitiated: false,
			},
		},
	}

	utils.SendJSONMessageMultiCast(readyMessage, room.Name)
}

func HandleReadyMessage(msg *ClientReadyMessage, usr *roomM.Connection) {

	if usr == nil {
		fmt.Println("Error: Connection not found for ready message")
		return
	}
	readyMessage(*msg, *usr)
}

func readyMessage(msg ClientReadyMessage, connection roomM.Connection) {

	// extract the data from the map
	isReady := msg.IsReady
	manuallyInitiated := msg.ManuallyInitiated

	// Get the room and user associated with the connection
	room := connection.Owner
	if room == nil {
		fmt.Println("Error: Room not found for connection")
		return
	}

	// Assuming username is extracted from the connection or another source
	if connection.Username == "" {
		fmt.Println("Error: Username not found for connection")
		return
	}

	// Set the user ready state
	room.SetUserReadyState(connection.Username, isReady, manuallyInitiated)

	// Send the ready message to all connections in the room
	readyMessage := map[string]interface{}{
		"Set": map[string]interface{}{
			"ready": map[string]interface{}{
				"username":          connection.Username,
				"isReady":           isReady,
				"manuallyInitiated": manuallyInitiated,
			},
		},
	}

	room.PrintReadyStates()

	// Send the ready message to all connections in the room

	utils.SendJSONMessageMultiCast(readyMessage, room.Name)
}
