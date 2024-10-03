package messages

import (
	"net"

	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type ReadyMessage struct {
	Set struct {
		Ready struct {
			Username          string      `json:"username"`
			IsReady           interface{} `json:"isReady"`
			ManuallyInitiated bool        `json:"manuallyInitiated"`
		} `json:"ready"`
	} `json:"Set"`
}

func SendReadyMessage(conn net.Conn, username string) {
	readyMessage := ReadyMessage{
		Set: struct {
			Ready struct {
				Username          string      `json:"username"`
				IsReady           interface{} `json:"isReady"`
				ManuallyInitiated bool        `json:"manuallyInitiated"`
			} `json:"ready"`
		}{
			Ready: struct {
				Username          string      `json:"username"`
				IsReady           interface{} `json:"isReady"`
				ManuallyInitiated bool        `json:"manuallyInitiated"`
			}{
				Username:          username,
				IsReady:           nil,
				ManuallyInitiated: false,
			},
		},
	}

	utils.SendJSONMessage(conn, readyMessage)
}
