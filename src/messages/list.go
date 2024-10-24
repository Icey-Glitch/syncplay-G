package messages

import (
	"fmt"
	"net"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

// FileInfo represents the file information.
type FileInfo struct {
	Duration float64 `json:"duration"`
	Name     string  `json:"name"`
	Size     float64 `json:"size"`
}

// PlayerInfo represents the player information.
type PlayerInfo struct {
	Position *float64 `json:"position,omitempty"`
	File     FileInfo `json:"file"`
}

// RoomInfo represents the room information.
type RoomInfo map[string]PlayerInfo

// ListResponse represents the list response.
type ListResponse struct {
	List map[string]RoomInfo `json:"List"`
}

// handleListRequest handles the "List" request and returns the response.
func HandleListRequest(conn net.Conn, room *roomM.Room) {
	// Retrieve user states from the PlaylistManager
	userStates, err := room.PlaylistManager.GetUsers()
	if err != true {
		fmt.Println("Error getting user states from the PlaylistManager")
		return
	}

	// Construct the players map
	players := make(map[string]PlayerInfo)
	for username, userState := range userStates {
		fileInfo := FileInfo{
			Duration: userState.Duration,
			Name:     userState.Name,
			Size:     userState.Size,
		}

		playerInfo := PlayerInfo{
			File: fileInfo,
		}

		// Add position if it exists
		if userState.Position != 0 {
			playerInfo.Position = &userState.Position
		}

		players[username] = playerInfo
	}

	// Retrieve the room name
	roomName := room.Name

	// Construct the room info
	roomInfo := RoomInfo(players)

	// Construct the list response
	listResponse := ListResponse{
		List: map[string]RoomInfo{
			roomName: roomInfo,
		},
	}

	// Send the response
	utils.SendJSONMessage(conn, listResponse, room.PlaylistManager, room.GetUsernameByConnection(conn))

}
