package messages

import (
	"net"

	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			Files []interface{} `json:"files"`
			User  interface{}   `json:"user"`
		} `json:"playlistChange"`
	} `json:"Set"`
}

type PlaylistIndexMessage struct {
	Set struct {
		PlaylistIndex struct {
			Index interface{} `json:"index"`
			User  interface{} `json:"user"`
		} `json:"playlistIndex"`
	} `json:"Set"`
}

func SendPlaylistIndexMessage(conn net.Conn) {
	playlistIndexMessage := PlaylistIndexMessage{
		Set: struct {
			PlaylistIndex struct {
				Index interface{} `json:"index"`
				User  interface{} `json:"user"`
			} `json:"playlistIndex"`
		}{
			PlaylistIndex: struct {
				Index interface{} `json:"index"`
				User  interface{} `json:"user"`
			}{
				Index: nil,
				User:  nil,
			},
		},
	}

	utils.SendJSONMessage(conn, playlistIndexMessage)
}

func SendPlaylistChangeMessage(conn net.Conn) {
	playlistChangeMessage := PlaylistChangeMessage{
		Set: struct {
			PlaylistChange struct {
				Files []interface{} `json:"files"`
				User  interface{}   `json:"user"`
			} `json:"playlistChange"`
		}{
			PlaylistChange: struct {
				Files []interface{} `json:"files"`
				User  interface{}   `json:"user"`
			}{
				Files: []interface{}{},
				User:  nil,
			},
		},
	}

	utils.SendJSONMessage(conn, playlistChangeMessage)
}

func ExtractStatePlaystateArguments(playstate map[string]interface{}) (interface{}, interface{}, interface{}, interface{}) {
	position, ok := playstate["position"].(float64)
	if !ok {
		position = 0
	}

	paused, ok := playstate["paused"].(bool)
	if !ok {
		paused = true
	}

	doSeek, ok := playstate["doSeek"].(bool)
	if !ok {
		doSeek = false
	}

	setBy := playstate["setBy"]

	return position, paused, doSeek, setBy
}
