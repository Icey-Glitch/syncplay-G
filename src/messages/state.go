package messages

import (
	"net"
	"time"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

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

type UserState struct {
	Position float64
	Paused   bool
	DoSeek   bool
	SetBy    interface{}
}

func SendInitialState(conn net.Conn, username string) {
	// check if the room exists
	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	if room == nil {
		return
	}

	// check if the room is empty
	if len(room.Users) == 0 {
		stateMessage := StateMessage{}
		stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9 // Convert to seconds
		stateMessage.State.Ping.ServerRtt = 0
		stateMessage.State.Playstate.DoSeek = false
		stateMessage.State.Playstate.Position = 0
		stateMessage.State.Playstate.Paused = true
		stateMessage.State.Playstate.SetBy = "Nobody" // Initial state, set by "Nobody"

		utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, username)
		return
	}

	// get the room's state
	roomState := room.RoomState
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9 // Convert to seconds
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = roomState.Position
	stateMessage.State.Playstate.Paused = roomState.IsPaused

	utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, username)

}

// add user to schedule
func SendUserState(room *roomM.Room, username string) bool {

	puser, exists := room.PlaylistManager.GetUserPlaystate(username)
	if !exists {
		return true
	}
	latencyCalculation := room.GetLatencyCalculation(username)

	conn := room.GetConnectionByUsername(username).Conn

	sendStateMessage(room, conn, puser.Position, puser.Paused, puser.DoSeek, latencyCalculation, puser.SetBy)
	return false
}

func sendStateMessage(room *roomM.Room, conn net.Conn, position, paused, doSeek, latencyCalculation, stateChange interface{}) {

	if room == nil {
		return
	}

	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = latencyCalculation.(float64)
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.Position = position.(float64)
	stateMessage.State.Playstate.Paused = paused.(bool)
	stateMessage.State.Playstate.DoSeek = doSeek.(bool)
	stateMessage.State.Playstate.SetBy = stateChange

	room.RoomState.Position = position.(float64)
	room.RoomState.IsPaused = paused.(bool)

	// update the room's state
	room.RoomState.Position = position.(float64)
	room.RoomState.IsPaused = paused.(bool)

	// send the state message to all users in the room

	utils.SendJSONMessage(conn, stateMessage, room.PlaylistManager, room.GetUsernameByConnection(conn))
}

func HandleStatePing(ping map[string]interface{}) (float64, float64) {
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

	messageAge := float64(time.Now().UnixNano()) / 1e9 // Convert to seconds
	latencyCalculation = (latencyCalculation + messageAge) / 2

	return messageAge, latencyCalculation
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
	room.PlaylistManager.SetUserPlaystate(room.GetUsernameByConnection(conn), position.(float64), paused.(bool), doSeek.(bool), setBy.(string), messageAge)

}

func GetLocalState() (interface{}, interface{}, interface{}, interface{}) {
	return globalState.position, globalState.paused, globalState.doSeek, globalState.setBy
}
