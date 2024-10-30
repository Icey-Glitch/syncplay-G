package messages

import (
	"fmt"
	"net"
	"strings"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			User  interface{}   `json:"user"`
			Files []interface{} `json:"files"`
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

// HandlePlaylistIndexMessage handle
func HandlePlaylistIndexMessage(connection roomM.Connection, playlistIndex map[string]interface{}) {

	room := connection.Owner
	if room == nil {
		return
	}

	playlistObject := room.PlaylistManager.GetPlaylist()

	if playlistObject.User.Username != "" {
		// check if the user is the same as the one who sent the message
		if playlistObject.User.Username != connection.Username {
			return
		}
	}

	index := playlistIndex["index"]

	if index != nil {
		playlistObject.Index = index.(float64)
	} else {
		playlistObject.Index = 0
	}

	playlistObject.User.Username = connection.Username

	room.PlaylistManager.SetPlaylist(playlistObject)

	if playlistObject.User.Username != connection.Username {
		return
	}
	SendPlaylistIndexMessage(connection)
}

func HandlePlaylistChangeMessage(connection roomM.Connection, playlistChange map[string]interface{}) {
	// client {"Set": {"playlistChange": {"files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}
	// server {"Set": {"playlistChange": {"user": "icey", "files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}

	room := connection.Owner
	if room == nil {
		return
	}

	PlaylistObject := room.PlaylistManager.GetPlaylist()

	fmt.Println("playlistChange: ", playlistChange) // map[files:[https://www.youtube.com/watch?v=0TVdTvWzr-A]]

	files, ok := playlistChange["files"].([]interface{})
	if !ok || files == nil {
		fmt.Println("Error: files is nil or not a slice of interfaces")
		return
	}

	// Sanitize URLs
	for i, file := range files {
		if url, ok := file.(string); ok {
			files[i] = strings.TrimSpace(url)
			fmt.Println("Sanitized URL: ", files[i])
		}
	}

	PlaylistObject.Files = files
	PlaylistObject.User.Username = connection.Username

	room.PlaylistManager.SetPlaylist(PlaylistObject)

	SendPlaylistChangeMessage(room, connection.Username)
}

// ExtractStatePlaystateArguments extract
func ExtractStatePlaystateArguments(playstate map[string]interface{}, connection roomM.Connection) (interface{}, interface{}, interface{}, interface{}) {
	if connection.Owner == nil {
		return nil, nil, nil, nil
	}

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
	if setBy == nil {
		setBy = connection.Username
	}

	return position, paused, doSeek, setBy
}

func SendPlaylistIndexMessage(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}

	PlaylistObject := connection.Owner.PlaylistManager.GetPlaylist()

	if PlaylistObject.User.Username == "" || len(PlaylistObject.Files) == 0 {
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
					Index: PlaylistObject.Index,
					User:  PlaylistObject.User.Username,
				},
			},
		}

		utils.SendJSONMessageMultiCast(playlistIndexMessage, connection.Owner.Name)
	} else {
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

		utils.SendJSONMessageMultiCast(playlistIndexMessage, connection.Owner.Name)
	}
}

func SendPlaylistChangeMessage(room *roomM.Room, username string) {
	if room == nil {
		return
	}
	PlaylistObject := room.PlaylistManager.GetPlaylist()

	fmt.Println("PlaylistObject: ", PlaylistObject)

	playlistChangeMessage := PlaylistChangeMessage{
		Set: struct {
			PlaylistChange struct {
				User  interface{}   `json:"user"`
				Files []interface{} `json:"files"`
			} `json:"playlistChange"`
		}{
			PlaylistChange: struct {
				User  interface{}   `json:"user"`
				Files []interface{} `json:"files"`
			}{
				Files: PlaylistObject.Files,
				User:  PlaylistObject.User.Username,
			},
		},
	}

	utils.SendJSONMessageMultiCast(playlistChangeMessage, room.Name)
}

func HandleFileMessage(conn net.Conn, file map[string]interface{}, Username string) {
	// Client >> {"Set": {"file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": 220514438}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": "220514438"}}}}}

	// Client >> {"Set": {"file": {"duration": 596.0, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "6fa13ad43fea"}, "file": {"duration": 596.458, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}}}

	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)
	if room == nil {
		fmt.Println("Error: room is nil")
		return
	}

	roomName := room.Name

	// desern communication type: raw, hashed, or not sent

	// extract the file data
	duration := file["duration"]
	name := file["name"]
	size := file["size"]

	// check if the file data is valid
	if duration == nil || name == nil || size == nil {
		fmt.Println("Error: file data is invalid")
		return
	}

	// store the user data
	err := room.PlaylistManager.SetUserFile(Username, duration.(float64), name.(string), size.(float64))
	if err != nil {
		fmt.Println("Error: failed to set user file data")
		return
	}

	// create the file message
	fileMessage := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				room.GetUsernameByConnection(conn): map[string]interface{}{
					"room": map[string]interface{}{
						"name": roomName,
					},
					"file": map[string]interface{}{
						"duration": duration,
						"name":     name,
						"size":     size,
					},
				},
			},
		},
	}

	// send the file message to all connections in the room
	utils.SendJSONMessageMultiCast(fileMessage, roomName)

}
