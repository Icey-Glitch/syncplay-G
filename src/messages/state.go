package messages

import (
	"net"
	"time"

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

func SendStateMessage(conn net.Conn, position, paused, doSeek, latencyCalculation, stateChange interface{}) {
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = latencyCalculation.(float64)
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = doSeek.(bool)
	stateMessage.State.Playstate.Position = position.(float64)
	stateMessage.State.Playstate.Paused = paused.(bool)
	stateMessage.State.Playstate.SetBy = stateChange

	utils.SendJSONMessage(conn, stateMessage)
}

func SendInitialState(conn net.Conn, username string) {
	stateMessage := StateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9 // Convert to seconds
	stateMessage.State.Ping.ServerRtt = 0
	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = 0
	stateMessage.State.Playstate.Paused = true
	stateMessage.State.Playstate.SetBy = nil // Initial state, no user set it

	utils.SendJSONMessage(conn, stateMessage)
}

func HandleStatePing(ping map[string]interface{}) (float64, float64) {
	messageAge, ok := ping["messageAge"].(float64)
	if !ok {
		messageAge = 0
	}
	latencyCalculation, ok := ping["latencyCalculation"].(float64)
	if !ok {
		latencyCalculation = 0
	}
	return messageAge, latencyCalculation
}

var globalState = struct {
	position float64
	paused   bool
	doSeek   bool
	setBy    interface{}
}{}

func UpdateGlobalState(position, paused, doSeek, setBy interface{}, messageAge float64) {
	globalState.position = position.(float64)
	globalState.paused = paused.(bool)
	globalState.doSeek = doSeek.(bool)
	globalState.setBy = setBy
}

func GetLocalState() (interface{}, interface{}, interface{}, interface{}) {
	return globalState.position, globalState.paused, globalState.doSeek, globalState.setBy
}
