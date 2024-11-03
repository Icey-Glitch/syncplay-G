package messages

import (
	"fmt"
	"strings"

	"github.com/Icey-Glitch/Syncplay-G/mngr/playlists"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			User  interface{}      `json:"user"`
			Files []playlists.File `json:"files"`
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

	// array of files to be added to the playlist
	FileObj := make([]playlists.File, len(files))
	size := 0.0
	duration := 0.0

	// Sanitize URLs
	for i, file := range files {
		if url, ok := file.(string); ok {
			files[i] = strings.TrimSpace(url)
			fmt.Println("Sanitized URL: ", files[i])
			// check if it is empty
			if files[i] != "" || files[i] != " " || files[i] != nil {
				FileObj[i] = playlists.File{
					Size:     size,
					Name:     url,
					Duration: duration,
				}
				room.PlaylistManager.AddFile(duration, url, size, connection.Username)
			}

		}
	}

	//room.PlaylistManager.AddFiles(FileObj, connection.Username)

	PlaylistObject.Files = connection.Owner.PlaylistManager.Playlist.Files
	PlaylistObject.User.Username = connection.Username

	SendPlaylistChangeMessage(connection)
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

func SendPlaylistChangeMessage(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}
	PlaylistObject := connection.Owner.PlaylistManager.GetPlaylist()

	fmt.Println("PlaylistObject: ", PlaylistObject)

	playlistChangeMessage := PlaylistChangeMessage{
		Set: struct {
			PlaylistChange struct {
				User  interface{}      `json:"user"`
				Files []playlists.File `json:"files"`
			} `json:"playlistChange"`
		}{
			PlaylistChange: struct {
				User  interface{}      `json:"user"`
				Files []playlists.File `json:"files"`
			}{
				Files: PlaylistObject.Files,
				User:  PlaylistObject.User.Username,
			},
		},
	}

	// if the files is nil return an empty array
	if playlistChangeMessage.Set.PlaylistChange.Files == nil {
		playlistChangeMessage.Set.PlaylistChange.Files = []playlists.File{}
	}

	utils.SendJSONMessageMultiCast(playlistChangeMessage, connection.Owner.Name)
}

func HandleFileMessage(connection roomM.Connection, file map[string]interface{}) {
	// Client >> {"Set": {"file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": 220514438}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "file": {"duration": 596.458, "name": "BigBuckBunny.avi", "size": "220514438"}}}}}

	// Client >> {"Set": {"file": {"duration": 596.0, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}
	// Server (to all who can see room) << {"Set": {"user": {"Bob": {"room": {"name": "6fa13ad43fea"}, "file": {"duration": 596.458, "name": "6fa13ad43fea", "size": "44657bd3c1bd"}}}}}

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
	fileObj := room.PlaylistManager.AddFile(duration.(float64), name.(string), size.(float64), connection.Username)
	room.PlaylistManager.SetUserFile(connection.Username, fileObj)

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
