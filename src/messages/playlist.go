package messages

import (
	"fmt"

	"github.com/Icey-Glitch/Syncplay-G/mngr/playlists"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type SetMessage struct {
	User           *UserMessage                 `json:"user,omitempty"`
	Ready          *ClientReadyMessage          `json:"ready,omitempty"`
	PlaylistChange *ClientPlaylistChangeMessage `json:"playlistChange,omitempty"`
	PlaylistIndex  *PlaylistIndexMessage        `json:"playlistIndex,omitempty"`
	File           *FileMessage                 `json:"file,omitempty"`
	Room           *RoomMessage                 `json:"room,omitempty"`
}

type PlaylistChangeMessage struct {
	Set struct {
		PlaylistChange struct {
			User  interface{} `json:"user"`
			Files []string    `json:"files"`
		} `json:"playlistChange"`
	} `json:"Set"`
}

type ClientPlaylistChangeMessage struct {
	Files []string `json:"files"`
}

type PlaylistIndexMessage struct {
	Set struct {
		PlaylistIndex struct {
			Index interface{} `json:"index"`
			User  string      `json:"user"`
		} `json:"playlistIndex"`
	} `json:"Set"`
}

type ClientPlaylistIndexMessage struct {
	Index int `json:"index"`
}

// HandlePlaylistIndexMessage handle
func HandlePlaylistIndexMessage(connection roomM.Connection, msg *PlaylistIndexMessage) {

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

	index := msg.Set.PlaylistIndex.Index

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

func HandlePlaylistChangeMessage(msg *ClientPlaylistChangeMessage, connection roomM.Connection) {
	// client {"Set": {"playlistChange": {"files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}
	// server {"Set": {"playlistChange": {"user": "icey", "files": ["https://www.youtube.com/watch?v=0TVdTvWzr-A"]}}}

	room := connection.Owner
	if room == nil {
		return
	}

	SendPlaylistChangeMessage(connection, msg.Files)
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

	playlistIndexMessage := PlaylistIndexMessage{}

	playlistIndexMessage.Set.PlaylistIndex.Index = PlaylistObject.Index
	playlistIndexMessage.Set.PlaylistIndex.User = connection.Username

	utils.SendJSONMessageMultiCast(playlistIndexMessage, connection.Owner.Name)

}

// SendPlaylistChangeMessage Takes in list of extracted files as a map of strings and then sends the message to all connections in the room
func SendPlaylistChangeMessage(connection roomM.Connection, files []string) {

	if connection.Owner == nil {
		return
	}
	PlaylistObject := connection.Owner.PlaylistManager.GetPlaylist()

	fmt.Println("PlaylistObject: ", PlaylistObject)

	playlistChangeMessage := PlaylistChangeMessage{}

	playlistChangeMessage.Set.PlaylistChange.User = connection.Username
	if files != nil {
		playlistChangeMessage.Set.PlaylistChange.Files = files
	} else {
		playlistChangeMessage.Set.PlaylistChange.Files = []string{}
	}

	fmt.Println("playlistChangeMessage: ", playlistChangeMessage)

	utils.SendJSONMessageMultiCast(playlistChangeMessage, connection.Owner.Name)
}

type FileMessage struct {
	Set struct {
		File struct {
			Duration float64     `json:"duration"`
			Name     string      `json:"name"`
			Size     interface{} `json:"size"`
		} `json:"file"`
	} `json:"Set"`
}

func HandleFileMessage(connection roomM.Connection, msg *FileMessage) {
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
	duration := msg.Set.File.Duration
	name := msg.Set.File.Name
	size := msg.Set.File.Size

	// check if the file data is valid
	// if duration < 0 || name == "" || size < 0 {
	// 	fmt.Println("Error: invalid file data")
	// 	return
	// }

	// check if size is sent hashed (not float64)
	var fileObj playlists.File
	var err error
	switch msg.Set.File.Size.(type) {
	case float64:
		fileObj, err = room.PlaylistManager.AddFile(duration, name, size.(float64), connection.Username, "")
		if err != nil {
			fmt.Println("Error: failed to add file to playlist")
			return
		}
	case nil:
		fileObj, err = room.PlaylistManager.AddFile(duration, name, 0, connection.Username, "")
		if err != nil {
			fmt.Println("Error: failed to add file to playlist")
			return
		}
	default:
		fileObj, err = room.PlaylistManager.AddFile(duration, name, 0, connection.Username, msg.Set.File.Size.(string))
		if err != nil {
			fmt.Println("Error: failed to add file to playlist")
			return
		}

	}

	err = room.PlaylistManager.SetUserFile(connection.Username, fileObj)
	if err != nil {
		fmt.Println("Error: failed to set user file")
		return
	}

	// create the file message
	fileMessage := FileMessage{}

	fileMessage.Set.File.Duration = duration
	fileMessage.Set.File.Name = name
	fileMessage.Set.File.Size = size

	// send the file message to all connections in the room
	utils.SendJSONMessageMultiCast(fileMessage, roomName)

}
