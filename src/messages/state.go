package messages

import (
	"fmt"
	"net"
	"time"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type StateMessage struct {
	State struct {
		Ping struct {
			LatencyCalculation       float64 `json:"latencyCalculation"`
			clientLatencyCalculation float64 `json:"clientLatencyCalculation"`
			ServerRtt                int     `json:"serverRtt"`
		} `json:"ping"`
		Playstate struct {
			Position float64     `json:"position"`
			Paused   bool        `json:"paused"`
			DoSeek   bool        `json:"doSeek"`
			SetBy    interface{} `json:"setBy"`
		} `json:"playstate"`
	} `json:"State"`
}

type UserState struct {
	Position float64
	Paused   bool
	DoSeek   bool
	SetBy    interface{}
}

func SendInitialState(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}

	if len(connection.Owner.Users) == 0 {
		stateMessage := StateMessage{}
		stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
		stateMessage.State.Ping.ServerRtt = 0
		stateMessage.State.Playstate.DoSeek = false
		stateMessage.State.Playstate.Position = 0
		stateMessage.State.Playstate.Paused = true
		stateMessage.State.Playstate.SetBy = "Nobody"

		err := utils.SendJSONMessage(connection.Conn, stateMessage, connection.Owner.PlaylistManager, connection.Username)
		if err != nil {
			fmt.Println("Error sending initial state message:", err)
			return
		}
		return
	}

	roomState := connection.Owner.RoomState
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = roomState.Position
	stateMessage.State.Playstate.Paused = roomState.IsPaused

	err := utils.SendJSONMessage(connection.Conn, stateMessage, connection.Owner.PlaylistManager, connection.Username)
	if err != nil {
		fmt.Println("Error sending initial state message:", err)
		return
	}
}

func SendUserState(connection roomM.Connection) bool {
	fmt.Println("Sending user state")
	puser, exists := connection.Owner.PlaylistManager.GetUserPlaystate(connection.Username)
	if !exists {
		fmt.Println("Error: User does not exist in the playlist:", connection.Username)
		return true
	}

	latencyCalculation, err := connection.Owner.GetUsersLatencyCalculation(&connection)
	if err != nil {
		fmt.Println("Error getting user latency calculation:", err)
		return true
	}

	processingTime := float64(time.Now().UnixNano())/1e9 - latencyCalculation.ArivalTime

	err = sendStateMessage(connection.Owner, connection.Conn, puser.Position, connection.Owner.PlaylistManager.Playlist.Paused, connection.Owner.PlaylistManager.Playlist.DoSeek, processingTime, puser.SetBy, latencyCalculation.ClientTime, connection.Username)
	if err != nil {
		fmt.Println("Error sending state message:", err)
		return true
	}

	return false
}

func sendStateMessage(room *roomM.Room, conn net.Conn, position float64, paused bool, doSeek bool, processingTime float64, stateChange string, clientTime float64, usr string) error {
	if room == nil {
		return fmt.Errorf("room cannot be nil")
	}

	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
	if clientTime != 0 {
		stateMessage.State.Ping.clientLatencyCalculation = clientTime + processingTime
	}
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.Position = position
	stateMessage.State.Playstate.Paused = paused
	stateMessage.State.Playstate.DoSeek = doSeek
	stateMessage.State.Playstate.SetBy = stateChange

	err := utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, usr)
	if err != nil {
		return fmt.Errorf("error sending JSON message: %w", err)
	}

	return nil
}

func HandleStatePing(ping map[string]interface{}) (clientRTT float64, latencyCalculation float64, clientLatencyCalculation float64) {
	/*
				General client ping (no file open / paused at beginning)

				Client >> {"State": {"ping": {"clientRtt": 0, "clientLatencyCalculation": 1394590473.585, "latencyCalculation": 1394590688.962084}, "playstate": {"paused": true, "position": 0.0}}}

				General client ping (file playing)

				Client >> {"State": {"ping": {"clientRtt": 0, "clientLatencyCalculation": 1394590473.585, "latencyCalculation": 1394590688.962084}, "playstate": {"paused": false, "position": 2.236}}}

				calculate the time the message was sent using the client's latency calculation and the server's time

				Look at the code to see how it works. ‘[client]LatencyCalculation’  is a timestamp based on the time in seconds since the epoch as a floating point number. ‘clientRtt’ is round-trip time (i.e. ping time). In older versions Syncplay used ‘yourLatency’ and ‘senderLatency’ but not ‘serverRtt’.

			{
		  "State": {
		    "ping": {
		      "clientRtt": 0,
		      "clientLatencyCalculation": 1394590473.585,
		      "latencyCalculation": 1394590688.962084
		    },
		    "playstate": {
		      "paused": true,
		      "position": 0
		    }
		  }
		}
	*/

	// TODO: Implement client latency calculation using last message time and current time (messageAge)
	latencyCalculation, ok := ping["latencyCalculation"].(float64)
	if !ok {
		latencyCalculation = 0
	}

	ClientRtt, ok := ping["clientRtt"].(float64)
	if !ok {
		ClientRtt = 0
	}

	ClientLatencyCalculation, ok := ping["clientLatencyCalculation"].(float64)
	if !ok {
		ClientLatencyCalculation = 0
	}

	// playstate logic

	return ClientRtt, latencyCalculation, ClientLatencyCalculation
}

var globalState = struct {
	position float64
	paused   bool
	doSeek   bool
	setBy    interface{}
}{}

func UpdateGlobalState(connection roomM.Connection, position, paused, doSeek, setBy interface{}, messageAge float64, latencyCalculation float64) {

	room := connection.Owner

	globalState.position = position.(float64)
	globalState.paused = paused.(bool)
	globalState.doSeek = doSeek.(bool)
	globalState.setBy = setBy

	error := room.PlaylistManager.SetUserPlaystate(connection.Username, position.(float64), paused.(bool), doSeek.(bool), setBy.(string), messageAge)
	if error != nil {
		fmt.Println("Error storing user playstate")
	}
}

func GetLocalState() (interface{}, interface{}, interface{}, interface{}) {
	return globalState.position, globalState.paused, globalState.doSeek, globalState.setBy
}
