package messages

import (
	"fmt"
	"net"
	"time"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
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

func SendInitialState(conn net.Conn, room *roomM.Room, username string) {
	if room == nil {
		return
	}

	if len(room.Users) == 0 {
		stateMessage := StateMessage{}
		stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
		stateMessage.State.Ping.ServerRtt = 0
		stateMessage.State.Playstate.DoSeek = false
		stateMessage.State.Playstate.Position = 0
		stateMessage.State.Playstate.Paused = true
		stateMessage.State.Playstate.SetBy = "Nobody"

		err := utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, username)
		if err != nil {
			fmt.Println("Error sending initial state message:", err)
			return
		}
		return
	}

	roomState := room.RoomState
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = roomState.Position
	stateMessage.State.Playstate.Paused = roomState.IsPaused

	err := utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, username)
	if err != nil {
		fmt.Println("Error sending initial state message:", err)
		return
	}
}

func SendUserState(room *roomM.Room, username string) bool {
	fmt.Println("Sending user state")
	connection := room.GetConnectionByUsername(username)
	if connection == nil {
		fmt.Println("Error: Connection not found for username:", username)
		return true
	}

	puser, exists := room.PlaylistManager.GetUserPlaystate(username)
	if !exists {
		fmt.Println("Error: User does not exist in the playlist:", username)
		return true
	}

	latencyCalculation, err := room.GetUsersLatencyCalculation(connection)
	if err != nil {
		fmt.Println("Error getting user latency calculation:", err)
		return true
	}

	processingTime := float64(time.Now().UnixNano())/1e9 - latencyCalculation.ArivalTime

	err = sendStateMessage(room, connection.Conn, puser.Position, puser.Paused, puser.DoSeek, processingTime, puser.SetBy, latencyCalculation.ClientTime, username)
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

	return ClientRtt, latencyCalculation, ClientLatencyCalculation
}

var globalState = struct {
	position float64
	paused   bool
	doSeek   bool
	setBy    interface{}
}{}

func UpdateGlobalState(conn net.Conn, position, paused, doSeek, setBy interface{}, messageAge float64, latencyCalculation float64) {
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)

	globalState.position = position.(float64)
	globalState.paused = paused.(bool)
	globalState.doSeek = doSeek.(bool)
	globalState.setBy = setBy

	// store the user's playstate
	error := room.PlaylistManager.SetUserPlaystate(room.GetUsernameByConnection(conn), position.(float64), paused.(bool), doSeek.(bool), setBy.(string), messageAge)
	if error != nil {
		fmt.Println("Error storing user playstate")
	}

}

func GetLocalState() (interface{}, interface{}, interface{}, interface{}) {
	return globalState.position, globalState.paused, globalState.doSeek, globalState.setBy
}
