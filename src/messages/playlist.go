package messages

import (
	"fmt"

	"github.com/Icey-Glitch/Syncplay-G/mngr/playlists"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			User  interface{} `json:"user"`
			Files interface{} `json:"files"`
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
func HandlePlaylistIndexMessage(connection roomM.Connection, value interface{}) {

	playlistIndex, ok := value.(map[string]interface{})
	if !ok || playlistIndex == nil {
		fmt.Println("Error: playlistIndex is nil or not a map")
		return
	}

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

func HandlePlaylistChangeMessage(value interface{}, connection roomM.Connection) {
	// client {"Set": {"playlistChange": {"files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}
	// server {"Set": {"playlistChange": {"user": "icey", "files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}

	playlistChange, ok := value.(map[string]interface{})
	if !ok || playlistChange == nil {
		fmt.Println("Error: playlistChange is nil or not a map")
		return
	}

	room := connection.Owner
	if room == nil {
		return
	}

	SendPlaylistChangeMessage(connection, playlistChange)
}

// ExtractStatePlaystateArguments extract
func ExtractStatePlaystateArguments(playstate map[string]interface{}, connection roomM.Connection) (interface{}, interface{}, interface{}, interface{}) {
	if connection.Owner == nil {
		return nil, nil, nil, nil
	}

	position, ok := playstate["position"].(float64)
	if !ok {
		position = 0.0
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
		setBy = "Nobody"
	}

	return position, paused, doSeek, setBy
}

func SendPlaylistIndexMessage(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}

	PlaylistObject := connection.Owner.PlaylistManager.GetPlaylist()

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

}

// SendPlaylistChangeMessage Takes in list of extracted files as a map of strings and then sends the message to all connections in the room
func SendPlaylistChangeMessage(connection roomM.Connection, files map[string]interface{}) {
	if files == nil {
		// empty files
		files = map[string]interface{}{
			"files": []string{},
		}
	}

	if connection.Owner == nil {
		return
	}
	PlaylistObject := connection.Owner.PlaylistManager.GetPlaylist()

	fmt.Println("PlaylistObject: ", PlaylistObject)

	playlistChangeMessage := PlaylistChangeMessage{
		Set: struct {
			PlaylistChange struct {
				User  interface{} `json:"user"`
				Files interface{} `json:"files"`
			} `json:"playlistChange"`
		}{
			PlaylistChange: struct {
				User  interface{} `json:"user"`
				Files interface{} `json:"files"`
			}{
				Files: files["files"],
				User:  PlaylistObject.User.Username,
			},
		},
	}

	// if the files is nil return an empty array
	if playlistChangeMessage.Set.PlaylistChange.Files == nil {
		playlistChangeMessage.Set.PlaylistChange.Files = []playlists.File{}
	}

	fmt.Println("playlistChangeMessage: ", playlistChangeMessage)

	utils.SendJSONMessageMultiCast(playlistChangeMessage, connection.Owner.Name)
}

func HandleFileMessage(connection roomM.Connection, value interface{}) {
	// Client >> {"Set": {"file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": 220514438}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": "220514438"}}}}}

	// Client >> {"Set": {"file": {"duration": 596.0, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "6fa13ad43fea"}, "file": {"duration": 596.458, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}}}

	file, ok := value.(map[string]interface{})
	if !ok || file == nil {
		fmt.Println("Error: file is nil or not a map")
		return
	}

	room := connection.Owner
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
	fileObj, err := room.PlaylistManager.AddFile(duration.(float64), name.(string), size.(float64), connection.Username)
	if err != nil {
		fmt.Println("Error: failed to add file to playlist")
		return
	}

	err = room.PlaylistManager.SetUserFile(connection.Username, fileObj)
	if err != nil {
		fmt.Println("Error: failed to set user file")
		return
	}

	// create the file message
	fileMessage := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				connection.Username: map[string]interface{}{
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
